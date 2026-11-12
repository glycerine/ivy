package cmd

func Cmd_FlagReadByName(f *Abc_Frame_t, name string) string {
	value, _ := f.LookupFlag(name)
	return value
}
func Cmd_FlagUpdateValue(f *Abc_Frame_t, name, value string) {
	if name != "" {
		f.SetFlag(name, value)
	}
}
func Cmd_FlagDeleteByName(f *Abc_Frame_t, name string) { f.DeleteFlag(name) }
