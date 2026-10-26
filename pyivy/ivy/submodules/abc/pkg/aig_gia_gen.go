package abc

// Packed simulation and dataset generation from src/aig/gia/giaGen.c.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
)

type GiaGenerationOptions struct {
	MaxObjects, MaxWords, MaxEntries int
	MaxSamples, MaxBytes             int
	Stop                             func() bool
}

type GiaGeneratedFile struct {
	Name string
	Data []byte
}

type GiaGeneratedDatasets struct {
	Files []GiaGeneratedFile
}

type GiaSimulationDataset struct {
	Inputs     int
	Words      int
	InputWords []uint64
	Outputs    []bool
}

type GiaSimulationComparison struct {
	Name    string `json:"name"`
	Inputs  int    `json:"input"`
	Outputs int    `json:"output"`
	ANDs    int    `json:"and"`
	Level   int    `json:"level"`
	Total   int    `json:"total"`
	Correct int    `json:"correct"`
	Guess   int    `json:"guess"`
}

func checkGiaGenerationOptions(options GiaGenerationOptions) (GiaGenerationOptions, error) {
	if options.MaxObjects < 0 || options.MaxWords < 0 || options.MaxEntries < 0 || options.MaxSamples < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 16
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	if options.MaxSamples == 0 {
		options.MaxSamples = 4 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	return options, nil
}

func checkGiaGenerationGraph(manager *GiaManager, options GiaGenerationOptions) error {
	if manager == nil {
		return ErrDimension
	}
	if err := manager.Check(); err != nil {
		return err
	}
	if manager.ObjectCount() > options.MaxObjects {
		return ErrLimit
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

func simulateGiaInputWords(manager *GiaManager, inputWords []uint64, words int, options GiaGenerationOptions) (GiaPackedSimulation, error) {
	if words < 0 || words > options.MaxWords || len(inputWords) != manager.InputCount()*words {
		return GiaPackedSimulation{}, ErrDimension
	}
	entries, err := checkedSize(0, manager.ObjectCount(), words, options.MaxEntries)
	if err != nil {
		return GiaPackedSimulation{}, err
	}
	result := GiaPackedSimulation{words: words, objects: manager.ObjectCount(), values: make([]uint64, entries)}
	for ordinal, id := range manager.cis {
		copy(result.values[int(id)*words:], inputWords[ordinal*words:(ordinal+1)*words])
	}
	read := func(literal GiaLit, word int) uint64 {
		value := result.values[int(literal.ID())*words+word]
		if literal.Complemented() {
			value = ^value
		}
		return value
	}
	for id, object := range manager.objects {
		if options.Stop != nil && options.Stop() {
			return GiaPackedSimulation{}, ErrCancelled
		}
		if object.IsAnd() {
			left, _ := object.Fanin0(GiaID(id))
			right, _ := object.Fanin1(GiaID(id))
			for word := 0; word < words; word++ {
				result.values[id*words+word] = read(left, word) & read(right, word)
			}
		} else if object.IsCO() {
			driver, _ := object.Fanin0(GiaID(id))
			for word := 0; word < words; word++ {
				result.values[id*words+word] = read(driver, word)
			}
		}
	}
	return result, nil
}

// SimulateInputWords propagates caller-owned CI-major words without modifying
// graph scratch. It represents Gia_ManSimulateWordsInit and the object helpers.
func (manager *GiaManager) SimulateInputWords(inputWords []uint64, options GiaPackedSimulationOptions) (GiaPackedSimulation, error) {
	if manager == nil || options.Words < 1 {
		return GiaPackedSimulation{}, ErrDimension
	}
	generation, err := checkGiaGenerationOptions(GiaGenerationOptions{MaxWords: options.MaxWords, MaxEntries: options.MaxEntries, Stop: options.Stop})
	if err != nil {
		return GiaPackedSimulation{}, err
	}
	if err = checkGiaGenerationGraph(manager, generation); err != nil {
		return GiaPackedSimulation{}, err
	}
	return simulateGiaInputWords(manager, slices.Clone(inputWords), options.Words, generation)
}

type giaGenerationRandom struct{ random ABCRandom }

func newGiaGenerationRandom(seed int) (*giaGenerationRandom, error) {
	if seed < 0 {
		return nil, ErrDimension
	}
	random := &giaGenerationRandom{}
	random.random.Uint64(true)
	for index := 0; index < seed; index++ {
		random.random.Uint64(false)
	}
	return random, nil
}

func (random *giaGenerationRandom) simulation(manager *GiaManager, words int, options GiaGenerationOptions) (GiaPackedSimulation, error) {
	if words < 0 || words > options.MaxWords {
		return GiaPackedSimulation{}, ErrLimit
	}
	count, err := checkedSize(0, manager.InputCount(), words, options.MaxEntries)
	if err != nil {
		return GiaPackedSimulation{}, err
	}
	inputs := make([]uint64, count)
	for index := range inputs {
		if options.Stop != nil && options.Stop() {
			return GiaPackedSimulation{}, ErrCancelled
		}
		inputs[index] = random.random.Uint64(false)
	}
	return simulateGiaInputWords(manager, inputs, words, options)
}

func giaGeneratedBase(manager *GiaManager, base string) string {
	if base != "" {
		return base
	}
	return manager.Name
}

func giaPackedDatasetFile(simulation GiaPackedSimulation, ids []GiaID, samples int, maximum int) ([]byte, error) {
	if samples < 0 || samples > math.MaxUint32 || len(ids) > math.MaxUint32 {
		return nil, ErrDimension
	}
	bits, err := checkedSize(0, samples, len(ids), math.MaxInt)
	if err != nil || bits%8 != 0 {
		return nil, ErrDimension
	}
	bytesCount := bits / 8
	if maximum < 12 || bytesCount > maximum-12 {
		return nil, ErrLimit
	}
	data := make([]byte, 12+bytesCount)
	binary.LittleEndian.PutUint32(data[0:], 2)
	binary.LittleEndian.PutUint32(data[4:], uint32(samples))
	binary.LittleEndian.PutUint32(data[8:], uint32(len(ids)))
	position := 0
	for sample := 0; sample < samples; sample++ {
		for _, id := range ids {
			bit, bitErr := simulation.Bit(id, sample)
			if bitErr != nil {
				return nil, bitErr
			}
			if bit {
				data[12+position/8] |= 1 << uint(position&7)
			}
			position++
		}
	}
	return data, nil
}

// GenerateGiaBinaryDatasets returns the four binary datasets and file-list
// artifact that Gia_ManDumpFiles wrote. Its source-compatible sample counts
// deliberately consume 64 simulation words per requested block but export
// only that many individual bits.
func GenerateGiaBinaryDatasets(manager *GiaManager, trainBlocks, testBlocks, seed int, base string, options GiaGenerationOptions) (GiaGeneratedDatasets, error) {
	options, err := checkGiaGenerationOptions(options)
	if err != nil || trainBlocks < 0 || testBlocks < 0 {
		return GiaGeneratedDatasets{}, ErrDimension
	}
	if err = checkGiaGenerationGraph(manager, options); err != nil {
		return GiaGeneratedDatasets{}, err
	}
	random, err := newGiaGenerationRandom(seed)
	if err != nil {
		return GiaGeneratedDatasets{}, err
	}
	name := giaGeneratedBase(manager, base)
	counts := []int{trainBlocks, testBlocks}
	suffixes := [][2]string{{"_x.train.data", "_y.train.data"}, {"_x.test.data", "_y.test.data"}}
	result := GiaGeneratedDatasets{}
	for dataset, blocks := range counts {
		if blocks > options.MaxWords/64 || blocks > options.MaxSamples/64 {
			return GiaGeneratedDatasets{}, ErrLimit
		}
		words := blocks * 64
		simulation, simulationErr := random.simulation(manager, words, options)
		if simulationErr != nil {
			return GiaGeneratedDatasets{}, simulationErr
		}
		inputs, inputErr := giaPackedDatasetFile(simulation, manager.cis, words, options.MaxBytes)
		outputs, outputErr := giaPackedDatasetFile(simulation, manager.cos, words, options.MaxBytes)
		if inputErr != nil {
			return GiaGeneratedDatasets{}, inputErr
		}
		if outputErr != nil {
			return GiaGeneratedDatasets{}, outputErr
		}
		result.Files = append(result.Files, GiaGeneratedFile{Name: name + suffixes[dataset][0], Data: inputs}, GiaGeneratedFile{Name: name + suffixes[dataset][1], Data: outputs})
	}
	var list strings.Builder
	for _, file := range result.Files {
		list.WriteString(file.Name)
		list.WriteByte('\n')
	}
	if list.Len() > options.MaxBytes {
		return GiaGeneratedDatasets{}, ErrLimit
	}
	result.Files = append(result.Files, GiaGeneratedFile{Name: name + ".flist", Data: []byte(list.String())})
	return result, nil
}

func giaSimulationPLA(manager *GiaManager, simulation GiaPackedSimulation, samples, maximum int) ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(&output, ".i %d\n.o %d\n.p %d\n.type fr\n", manager.InputCount(), manager.OutputCount(), samples)
	for sample := 0; sample < samples; sample++ {
		for _, id := range manager.cis {
			bit, err := simulation.Bit(id, sample)
			if err != nil {
				return nil, err
			}
			output.WriteByte("01"[boolIndex(bit)])
		}
		output.WriteByte(' ')
		for _, id := range manager.cos {
			bit, err := simulation.Bit(id, sample)
			if err != nil {
				return nil, err
			}
			output.WriteByte("01"[boolIndex(bit)])
		}
		output.WriteByte('\n')
		if output.Len() > maximum {
			return nil, ErrLimit
		}
	}
	output.WriteString(".e\n")
	if output.Len() > maximum {
		return nil, ErrLimit
	}
	return output.Bytes(), nil
}

func boolIndex(value bool) int {
	if value {
		return 1
	}
	return 0
}

// GenerateGiaPLADatasets returns train, validation and test PLAs without
// touching the filesystem. Validation and test have equal source counts but
// consume consecutive random words.
func GenerateGiaPLADatasets(manager *GiaManager, trainWords, validationWords, seed int, base string, options GiaGenerationOptions) (GiaGeneratedDatasets, error) {
	options, err := checkGiaGenerationOptions(options)
	if err != nil || trainWords < 0 || validationWords < 0 {
		return GiaGeneratedDatasets{}, ErrDimension
	}
	if err = checkGiaGenerationGraph(manager, options); err != nil {
		return GiaGeneratedDatasets{}, err
	}
	random, err := newGiaGenerationRandom(seed)
	if err != nil {
		return GiaGeneratedDatasets{}, err
	}
	name := giaGeneratedBase(manager, base)
	counts := []int{trainWords, validationWords, validationWords}
	suffixes := []string{".train.pla", ".valid.pla", ".test.pla"}
	result := GiaGeneratedDatasets{}
	for index, words := range counts {
		if words > options.MaxWords || words > options.MaxSamples/64 {
			return GiaGeneratedDatasets{}, ErrLimit
		}
		simulation, simulationErr := random.simulation(manager, words, options)
		if simulationErr != nil {
			return GiaGeneratedDatasets{}, simulationErr
		}
		data, encodeErr := giaSimulationPLA(manager, simulation, words*64, options.MaxBytes)
		if encodeErr != nil {
			return GiaGeneratedDatasets{}, encodeErr
		}
		result.Files = append(result.Files, GiaGeneratedFile{Name: name + suffixes[index], Data: data})
	}
	return result, nil
}

func ParseGiaSimulationPLA(data []byte, options GiaGenerationOptions) (GiaSimulationDataset, error) {
	options, err := checkGiaGenerationOptions(options)
	if err != nil {
		return GiaSimulationDataset{}, err
	}
	if len(data) > options.MaxBytes {
		return GiaSimulationDataset{}, ErrLimit
	}
	rows := make([][]byte, 0)
	for _, raw := range bytes.Split(data, []byte{'\n'}) {
		if options.Stop != nil && options.Stop() {
			return GiaSimulationDataset{}, ErrCancelled
		}
		if bytes.Contains(raw, []byte{'.'}) {
			continue
		}
		bits := make([]byte, 0, len(raw))
		for _, value := range raw {
			if value == '0' || value == '1' {
				bits = append(bits, value)
			}
		}
		if len(bits) != 0 {
			rows = append(rows, bits)
		}
	}
	if len(rows) == 0 || len(rows)%64 != 0 || len(rows) > options.MaxSamples || len(rows[0]) < 1 {
		return GiaSimulationDataset{}, ErrDimension
	}
	width := len(rows[0])
	for _, row := range rows {
		if len(row) != width {
			return GiaSimulationDataset{}, ErrDimension
		}
	}
	inputs, words := width-1, len(rows)/64
	entries, err := checkedSize(0, inputs, words, options.MaxEntries)
	if err != nil || words > options.MaxWords {
		return GiaSimulationDataset{}, ErrLimit
	}
	result := GiaSimulationDataset{Inputs: inputs, Words: words, InputWords: make([]uint64, entries), Outputs: make([]bool, len(rows))}
	for sample, row := range rows {
		for input := 0; input < inputs; input++ {
			if row[input] == '1' {
				result.InputWords[input*words+sample/64] |= uint64(1) << uint(sample&63)
			}
		}
		result.Outputs[sample] = row[inputs] == '1'
	}
	return result, nil
}

func (manager *GiaManager) CompareSimulationDataset(dataset GiaSimulationDataset, options GiaGenerationOptions) (GiaSimulationComparison, error) {
	options, err := checkGiaGenerationOptions(options)
	if err != nil {
		return GiaSimulationComparison{}, err
	}
	if err = checkGiaGenerationGraph(manager, options); err != nil {
		return GiaSimulationComparison{}, err
	}
	if dataset.Inputs != manager.InputCount() || dataset.Words < 1 || dataset.Words > options.MaxWords || len(dataset.InputWords) != dataset.Inputs*dataset.Words || len(dataset.Outputs) != 64*dataset.Words || len(dataset.Outputs) > options.MaxSamples || manager.OutputCount() == 0 {
		return GiaSimulationComparison{}, ErrDimension
	}
	simulation, err := simulateGiaInputWords(manager, slices.Clone(dataset.InputWords), dataset.Words, options)
	if err != nil {
		return GiaSimulationComparison{}, err
	}
	correct, ones := 0, 0
	for sample, expected := range dataset.Outputs {
		actual, bitErr := simulation.Bit(manager.cos[0], sample)
		if bitErr != nil {
			return GiaSimulationComparison{}, bitErr
		}
		if actual == expected {
			correct++
		}
		if expected {
			ones++
		}
	}
	_, level, err := manager.Levels(options.MaxObjects)
	if err != nil {
		return GiaSimulationComparison{}, err
	}
	total := len(dataset.Outputs)
	guess := total - ones
	if ones > total/2 {
		guess = ones
	}
	return GiaSimulationComparison{Name: manager.Name, Inputs: manager.InputCount(), Outputs: manager.OutputCount(), ANDs: manager.AndCount(), Level: int(level), Total: total, Correct: correct, Guess: guess}, nil
}

func (comparison GiaSimulationComparison) MarshalStatsJSON(maximum int) ([]byte, error) {
	if maximum < 0 {
		return nil, ErrDimension
	}
	data, err := json.MarshalIndent(comparison, "", "    ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if maximum > 0 && len(data) > maximum {
		return nil, ErrLimit
	}
	return data, nil
}

func (manager *GiaManager) TestSimulationPLA(data []byte, options GiaGenerationOptions) (GiaSimulationComparison, error) {
	dataset, err := ParseGiaSimulationPLA(data, options)
	if err != nil {
		return GiaSimulationComparison{}, err
	}
	return manager.CompareSimulationDataset(dataset, options)
}
