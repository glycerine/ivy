#include "Vmemory_system.h"
#include "verilated.h"

#include <array>
#include <cstdint>
#include <exception>
#include <iostream>
#include <optional>
#include <stdexcept>
#include <unordered_map>

struct Request {
    uint64_t line;
    unsigned offset;
    unsigned slot;
    bool store;
    uint64_t value;
    bool matched_previous = false;
    bool accepted = false;
};

struct CpuPorts {
    CData &enable;
    QData &line;
    CData &offset;
    CData &store;
    QData &write_data;
    CData &load_slot;
    CData &store_slot;
    const CData &ready;
    const CData &read_valid;
    const QData &read_line;
    const CData &read_offset;
    const CData &read_slot;
    const QData &read_data;
    const CData &ack_valid;
    const CData &ack_slot;
    const QData &ack_data;
};

static uint64_t initial_word(uint64_t line, unsigned offset) {
    return UINT64_C(0xDEAD0123456789AB) ^ (line * 1125899906842597 + offset);
}

class CacheTest {
public:
    CacheTest() {
        dut.posedge = 0;
        dut.rst = 1;
        dut.cpu0_req_en = 0;
        dut.cpu1_req_en = 0;
        for (int i = 0; i < 4; ++i)
            tick();
        dut.rst = 0;
        dut.eval();
    }

    void run(std::optional<Request> cpu0, std::optional<Request> cpu1) {
        std::array<std::optional<Request>, 2> requests{cpu0, cpu1};
        for (int cycle = 0; cycle < 600; ++cycle) {
            if (!requests[0] && !requests[1])
                return;
            for (unsigned cpu = 0; cpu < 2; ++cpu)
                drive(ports(cpu), requests[cpu]);
            prepare();
            for (unsigned cpu = 0; cpu < 2; ++cpu)
                observe(cpu, ports(cpu), requests[cpu]);
            edge();
        }
        throw std::runtime_error("cache request failed to complete within 600 cycles");
    }

private:
    Vmemory_system dut;
    using Beat = std::array<uint32_t, 8>;
    std::unordered_map<uint64_t, Beat> memory;

    CpuPorts ports(unsigned cpu) {
        if (cpu == 0)
            return {dut.cpu0_req_en, dut.cpu0_req_h, dut.cpu0_req_o, dut.cpu0_req_wen,
                    dut.cpu0_req_wdata, dut.cpu0_load_slot, dut.cpu0_store_slot, dut.cpu0_ready,
                    dut.cpu0_resp_valid, dut.cpu0_resp_h, dut.cpu0_resp_o, dut.cpu0_resp_idx,
                    dut.cpu0_resp_data, dut.cpu0_store_ack_valid, dut.cpu0_store_ack_idx,
                    dut.cpu0_store_ack_data};
        return {dut.cpu1_req_en, dut.cpu1_req_h, dut.cpu1_req_o, dut.cpu1_req_wen,
                dut.cpu1_req_wdata, dut.cpu1_load_slot, dut.cpu1_store_slot, dut.cpu1_ready,
                dut.cpu1_resp_valid, dut.cpu1_resp_h, dut.cpu1_resp_o, dut.cpu1_resp_idx,
                dut.cpu1_resp_data, dut.cpu1_store_ack_valid, dut.cpu1_store_ack_idx,
                dut.cpu1_store_ack_data};
    }

    static void drive(CpuPorts pins, const std::optional<Request> &request) {
        pins.enable = request && !request->accepted;
        if (!pins.enable)
            return;
        pins.line = request->line;
        pins.offset = request->offset;
        pins.store = request->store;
        pins.write_data = request->value;
        pins.load_slot = pins.store_slot = request->slot;
    }

    static void observe(unsigned cpu, const CpuPorts &pins, std::optional<Request> &pending) {
        if (!pending) {
            if (pins.ack_valid || pins.read_valid)
                throw std::runtime_error("unsolicited CPU response");
            return;
        }
        Request &request = *pending;
        if (pins.enable && pins.ready)
            request.accepted = true;
        if (!pins.read_valid && !pins.ack_valid)
            return;
        if (!request.accepted || pins.read_valid == pins.ack_valid ||
            bool(pins.ack_valid) != request.store)
            throw std::runtime_error("response did not match an accepted request");
        if (request.store) {
            if (pins.ack_slot != request.slot || pins.ack_data != uint64_t(request.matched_previous))
                throw std::runtime_error("incorrect store acknowledgement");
        } else if (pins.read_slot != request.slot || pins.read_line != request.line ||
                   pins.read_offset != request.offset || pins.read_data != request.value) {
            std::cerr << "CPU " << cpu << " read " << std::hex << pins.read_data
                      << " from line " << pins.read_line << ", expected " << request.value
                      << " from line " << request.line << '\n';
            throw std::runtime_error("incorrect load response");
        }
        pending.reset();
    }

    Beat read_beat(uint64_t address) const {
        const uint64_t aligned = address & ~UINT64_C(3);
        const auto saved = memory.find(aligned);
        if (saved != memory.end())
            return saved->second;
        Beat initial{};
        for (unsigned i = 0; i < 4; ++i) {
            uint64_t value = initial_word(aligned >> 3, (aligned & 7) + i);
            initial[2 * i] = uint32_t(value);
            initial[2 * i + 1] = value >> 32;
        }
        return initial;
    }

    void prepare() {
        dut.posedge = 0;
        dut.eval();
        const Beat input = dut.mem_rd_en ? read_beat(dut.mem_rd_addr) : Beat{};
        for (unsigned i = 0; i < input.size(); ++i)
            dut.mem_rdata[i] = input[i];
        dut.eval();
    }

    void edge() {
        if (!dut.rst && dut.mem_wr_en) {
            Beat &data = memory[dut.mem_wr_addr & ~UINT64_C(3)];
            for (unsigned i = 0; i < data.size(); ++i)
                data[i] = dut.mem_wr_data[i];
        }
        dut.posedge = 1;
        dut.eval();
    }

    void tick() {
        prepare();
        edge();
    }
};

int main(int argc, char **argv) {
    Verilated::commandArgs(argc, argv);
    try {
        constexpr uint64_t first = 0x34, second = 0x51;
        constexpr uint64_t written_first = UINT64_C(0x1234567887654321);
        constexpr uint64_t written_second = UINT64_C(0x0123456789ABCDEF);
        CacheTest cache;
        cache.run(Request{first, 0, 1, false, initial_word(first, 0)},
                  Request{second, 6, 2, false, initial_word(second, 6)});
        cache.run(Request{first, 0, 3, true, written_first},
                  Request{second, 6, 4, true, written_second});
        cache.run(Request{second, 6, 5, false, written_second},
                  Request{first, 0, 6, false, written_first});
        cache.run(Request{second, 1, 7, false, initial_word(second, 1)},
                  Request{first, 5, 0, false, initial_word(first, 5)});
        cache.run(Request{first, 0, 1, true, written_first, true}, std::nullopt);
        std::cout << "German cache read, store, and coherence scenarios passed\n";
        return 0;
    } catch (const std::exception &error) {
        std::cerr << error.what() << '\n';
        return 1;
    }
}
