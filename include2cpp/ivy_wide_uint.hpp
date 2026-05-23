#pragma once

#include <algorithm>
#include <climits>
#include <cstddef>
#include <cstdint>
#include <cstdlib>
#include <iostream>
#include <stdexcept>
#include <string>

#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)

static inline unsigned __int128 ivy_uint128_mask(unsigned bits) {
    if (bits == 0) {
        return 0;
    }
    if (bits >= 128) {
        return ~((unsigned __int128)0);
    }
    return (((unsigned __int128)1) << bits) - 1;
}

static inline std::string ivy_uint128_to_string(unsigned __int128 value) {
    if (value == 0) {
        return "0";
    }
    std::string out;
    while (value != 0) {
        unsigned digit = static_cast<unsigned>(value % 10);
        out.push_back(static_cast<char>('0' + digit));
        value /= 10;
    }
    std::reverse(out.begin(), out.end());
    return out;
}

static inline int ivy_uint_digit_value(char c) {
    if (c >= '0' && c <= '9') {
        return c - '0';
    }
    if (c >= 'a' && c <= 'f') {
        return c - 'a' + 10;
    }
    if (c >= 'A' && c <= 'F') {
        return c - 'A' + 10;
    }
    return -1;
}

static inline unsigned __int128 ivy_uint128_from_string(const std::string &text) {
    std::size_t pos = 0;
    while (pos < text.size() && (text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r')) {
        pos++;
    }
    bool neg = false;
    if (pos < text.size() && (text[pos] == '+' || text[pos] == '-')) {
        neg = text[pos] == '-';
        pos++;
    }
    unsigned base = 10;
    if (pos + 1 < text.size() && text[pos] == '0') {
        if (text[pos + 1] == 'x' || text[pos + 1] == 'X') {
            base = 16;
            pos += 2;
        } else if (text[pos + 1] == 'b' || text[pos + 1] == 'B') {
            base = 2;
            pos += 2;
        } else {
            base = 8;
            pos++;
        }
    }
    unsigned __int128 value = 0;
    for (; pos < text.size(); pos++) {
        char c = text[pos];
        if (c == '_') {
            continue;
        }
        int digit = ivy_uint_digit_value(c);
        if (digit < 0 || static_cast<unsigned>(digit) >= base) {
            break;
        }
        value = value * base + static_cast<unsigned>(digit);
    }
    return neg ? (0 - value) : value;
}

static inline unsigned __int128 ivy_uint128_random(unsigned bits) {
    unsigned __int128 value = 0;
    for (unsigned i = 0; i < 5; i++) {
        value <<= 31;
        value ^= static_cast<unsigned>(std::rand() & 0x7fffffff);
    }
    return value & ivy_uint128_mask(bits);
}

inline std::ostream &operator<<(std::ostream &out, unsigned __int128 value) {
    out << ivy_uint128_to_string(value);
    return out;
}

inline std::istream &operator>>(std::istream &in, unsigned __int128 &value) {
    std::string text;
    in >> text;
    value = ivy_uint128_from_string(text);
    return in;
}

template <unsigned Bits>
class ivy_uint {
public:
    static const unsigned bit_count = Bits;
    static const unsigned word_bits = 64;
    static const unsigned word_count = (Bits + word_bits - 1) / word_bits;

    unsigned long long words[word_count];

    ivy_uint() {
        clear();
    }

    ivy_uint(unsigned long long value) {
        clear();
        words[0] = value;
        normalize();
    }

    ivy_uint(unsigned value) {
        clear();
        words[0] = value;
        normalize();
    }

    ivy_uint(int value) {
        clear();
        words[0] = static_cast<unsigned long long>(value);
        normalize();
    }

    ivy_uint(long long value) {
        clear();
        words[0] = static_cast<unsigned long long>(value);
        normalize();
    }

#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)
    ivy_uint(unsigned __int128 value) {
        clear();
        words[0] = static_cast<unsigned long long>(value);
        if (word_count > 1) {
            words[1] = static_cast<unsigned long long>(value >> 64);
        }
        normalize();
    }
#endif

    template <unsigned OtherBits>
    ivy_uint(const ivy_uint<OtherBits> &other) {
        clear();
        const unsigned n = word_count < ivy_uint<OtherBits>::word_count ? word_count : ivy_uint<OtherBits>::word_count;
        for (unsigned i = 0; i < n; i++) {
            words[i] = other.words[i];
        }
        normalize();
    }

    explicit ivy_uint(const std::string &text) {
        clear();
        from_string(text);
    }

    static ivy_uint mask() {
        ivy_uint res;
        for (unsigned i = 0; i < word_count; i++) {
            res.words[i] = ~0ULL;
        }
        res.normalize();
        return res;
    }

    static ivy_uint random() {
        ivy_uint res;
        for (unsigned i = 0; i < word_count; i++) {
            unsigned long long w = 0;
            for (unsigned j = 0; j < 3; j++) {
                w <<= 21;
                w ^= static_cast<unsigned long long>(std::rand() & 0x1fffff);
            }
            w <<= 1;
            w ^= static_cast<unsigned long long>(std::rand() & 1);
            res.words[i] = w;
        }
        res.normalize();
        return res;
    }

    void clear() {
        for (unsigned i = 0; i < word_count; i++) {
            words[i] = 0;
        }
    }

    void normalize() {
        words[word_count - 1] &= top_mask();
    }

    bool is_zero() const {
        for (unsigned i = 0; i < word_count; i++) {
            if (words[i] != 0) {
                return false;
            }
        }
        return true;
    }

    bool bit(unsigned idx) const {
        if (idx >= Bits) {
            return false;
        }
        return (words[idx / word_bits] >> (idx % word_bits)) & 1ULL;
    }

    void set_bit(unsigned idx) {
        if (idx < Bits) {
            words[idx / word_bits] |= (1ULL << (idx % word_bits));
        }
    }

    std::string to_decimal_string() const {
        if (is_zero()) {
            return "0";
        }
        ivy_uint tmp(*this);
        std::string out;
        while (!tmp.is_zero()) {
            unsigned digit = tmp.div_small(10);
            out.push_back(static_cast<char>('0' + digit));
        }
        std::reverse(out.begin(), out.end());
        return out;
    }

    void from_string(const std::string &text) {
        clear();
        std::size_t pos = 0;
        while (pos < text.size() && (text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r')) {
            pos++;
        }
        bool neg = false;
        if (pos < text.size() && (text[pos] == '+' || text[pos] == '-')) {
            neg = text[pos] == '-';
            pos++;
        }
        unsigned base = 10;
        if (pos + 1 < text.size() && text[pos] == '0') {
            if (text[pos + 1] == 'x' || text[pos + 1] == 'X') {
                base = 16;
                pos += 2;
            } else if (text[pos + 1] == 'b' || text[pos + 1] == 'B') {
                base = 2;
                pos += 2;
            } else {
                base = 8;
                pos++;
            }
        }
        for (; pos < text.size(); pos++) {
            char c = text[pos];
            if (c == '_') {
                continue;
            }
            int digit = ivy_uint_digit_value(c);
            if (digit < 0 || static_cast<unsigned>(digit) >= base) {
                break;
            }
            mul_small(base);
            add_small(static_cast<unsigned>(digit));
        }
        normalize();
        if (neg) {
            ivy_uint zero;
            *this = zero - *this;
        }
    }

    size_t __hash() const {
        size_t res = 0;
        for (unsigned i = 0; i < word_count; i++) {
            res ^= static_cast<size_t>(words[i] + 0x9e3779b97f4a7c15ULL + (res << 6) + (res >> 2));
        }
        return res;
    }

    explicit operator unsigned() const {
        return static_cast<unsigned>(words[0]);
    }

    explicit operator unsigned long long() const {
        return words[0];
    }

#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)
    explicit operator unsigned __int128() const {
        unsigned __int128 res = words[0];
        if (word_count > 1) {
            res |= static_cast<unsigned __int128>(words[1]) << 64;
        }
        return res;
    }
#endif

    unsigned to_shift_amount() const {
        if (word_count > 1) {
            for (unsigned i = 1; i < word_count; i++) {
                if (words[i] != 0) {
                    return UINT_MAX;
                }
            }
        }
        if (words[0] > static_cast<unsigned long long>(UINT_MAX)) {
            return UINT_MAX;
        }
        return static_cast<unsigned>(words[0]);
    }

    ivy_uint &operator++() {
        *this = *this + ivy_uint(1);
        return *this;
    }

    ivy_uint operator++(int) {
        ivy_uint old(*this);
        ++(*this);
        return old;
    }

    ivy_uint &operator+=(const ivy_uint &other) {
        *this = *this + other;
        return *this;
    }

    ivy_uint &operator-=(const ivy_uint &other) {
        *this = *this - other;
        return *this;
    }

    ivy_uint &operator*=(const ivy_uint &other) {
        *this = *this * other;
        return *this;
    }

    ivy_uint &operator&=(const ivy_uint &other) {
        *this = *this & other;
        return *this;
    }

    ivy_uint &operator|=(const ivy_uint &other) {
        *this = *this | other;
        return *this;
    }

    ivy_uint &operator^=(const ivy_uint &other) {
        *this = *this ^ other;
        return *this;
    }

    ivy_uint &operator<<=(unsigned shift) {
        *this = *this << shift;
        return *this;
    }

    ivy_uint &operator>>=(unsigned shift) {
        *this = *this >> shift;
        return *this;
    }

    friend bool operator==(const ivy_uint &a, const ivy_uint &b) {
        for (unsigned i = 0; i < word_count; i++) {
            if (a.words[i] != b.words[i]) {
                return false;
            }
        }
        return true;
    }

    friend bool operator!=(const ivy_uint &a, const ivy_uint &b) {
        return !(a == b);
    }

    friend bool operator<(const ivy_uint &a, const ivy_uint &b) {
        for (unsigned i = word_count; i-- > 0;) {
            if (a.words[i] < b.words[i]) {
                return true;
            }
            if (a.words[i] > b.words[i]) {
                return false;
            }
        }
        return false;
    }

    friend bool operator>(const ivy_uint &a, const ivy_uint &b) {
        return b < a;
    }

    friend bool operator<=(const ivy_uint &a, const ivy_uint &b) {
        return !(b < a);
    }

    friend bool operator>=(const ivy_uint &a, const ivy_uint &b) {
        return !(a < b);
    }

    friend ivy_uint operator&(ivy_uint a, const ivy_uint &b) {
        for (unsigned i = 0; i < word_count; i++) {
            a.words[i] &= b.words[i];
        }
        return a;
    }

    friend ivy_uint operator|(ivy_uint a, const ivy_uint &b) {
        for (unsigned i = 0; i < word_count; i++) {
            a.words[i] |= b.words[i];
        }
        a.normalize();
        return a;
    }

    friend ivy_uint operator^(ivy_uint a, const ivy_uint &b) {
        for (unsigned i = 0; i < word_count; i++) {
            a.words[i] ^= b.words[i];
        }
        a.normalize();
        return a;
    }

    friend ivy_uint operator~(ivy_uint a) {
        for (unsigned i = 0; i < word_count; i++) {
            a.words[i] = ~a.words[i];
        }
        a.normalize();
        return a;
    }

    friend ivy_uint operator-(const ivy_uint &a) {
        return ivy_uint(static_cast<unsigned long long>(0)) - a;
    }

    friend ivy_uint operator+(ivy_uint a, const ivy_uint &b) {
        unsigned __int128 carry = 0;
        for (unsigned i = 0; i < word_count; i++) {
            unsigned __int128 sum = static_cast<unsigned __int128>(a.words[i]) + b.words[i] + carry;
            a.words[i] = static_cast<unsigned long long>(sum);
            carry = sum >> 64;
        }
        a.normalize();
        return a;
    }

    friend ivy_uint operator-(ivy_uint a, const ivy_uint &b) {
        unsigned __int128 borrow = 0;
        for (unsigned i = 0; i < word_count; i++) {
            unsigned __int128 lhs = static_cast<unsigned __int128>(a.words[i]);
            unsigned __int128 rhs = static_cast<unsigned __int128>(b.words[i]) + borrow;
            a.words[i] = static_cast<unsigned long long>(lhs - rhs);
            borrow = lhs < rhs ? 1 : 0;
        }
        a.normalize();
        return a;
    }

    friend ivy_uint operator*(const ivy_uint &a, const ivy_uint &b) {
        ivy_uint res;
        for (unsigned i = 0; i < word_count; i++) {
            unsigned __int128 carry = 0;
            for (unsigned j = 0; i + j < word_count; j++) {
                unsigned __int128 cur = static_cast<unsigned __int128>(a.words[i]) * b.words[j];
                cur += res.words[i + j];
                cur += carry;
                res.words[i + j] = static_cast<unsigned long long>(cur);
                carry = cur >> 64;
            }
        }
        res.normalize();
        return res;
    }

    friend ivy_uint operator<<(const ivy_uint &a, unsigned shift) {
        ivy_uint res;
        if (shift >= Bits) {
            return res;
        }
        unsigned word_shift = shift / word_bits;
        unsigned bit_shift = shift % word_bits;
        for (unsigned i = word_count; i-- > word_shift;) {
            unsigned long long value = a.words[i - word_shift] << bit_shift;
            if (bit_shift != 0 && i > word_shift) {
                value |= a.words[i - word_shift - 1] >> (word_bits - bit_shift);
            }
            res.words[i] = value;
        }
        res.normalize();
        return res;
    }

    friend ivy_uint operator<<(const ivy_uint &a, int shift) {
        return a << static_cast<unsigned>(shift);
    }

    friend ivy_uint operator>>(const ivy_uint &a, unsigned shift) {
        ivy_uint res;
        if (shift >= Bits) {
            return res;
        }
        unsigned word_shift = shift / word_bits;
        unsigned bit_shift = shift % word_bits;
        for (unsigned i = 0; i + word_shift < word_count; i++) {
            unsigned long long value = a.words[i + word_shift] >> bit_shift;
            if (bit_shift != 0 && i + word_shift + 1 < word_count) {
                value |= a.words[i + word_shift + 1] << (word_bits - bit_shift);
            }
            res.words[i] = value;
        }
        res.normalize();
        return res;
    }

    friend ivy_uint operator>>(const ivy_uint &a, int shift) {
        return a >> static_cast<unsigned>(shift);
    }

    friend ivy_uint operator/(const ivy_uint &a, const ivy_uint &b) {
        ivy_uint q;
        ivy_uint r;
        divmod(a, b, q, r);
        return q;
    }

    friend ivy_uint operator%(const ivy_uint &a, const ivy_uint &b) {
        ivy_uint q;
        ivy_uint r;
        divmod(a, b, q, r);
        return r;
    }

    friend std::ostream &operator<<(std::ostream &out, const ivy_uint &value) {
        out << value.to_decimal_string();
        return out;
    }

    friend std::istream &operator>>(std::istream &in, ivy_uint &value) {
        std::string text;
        in >> text;
        value = ivy_uint(text);
        return in;
    }

private:
    static unsigned long long top_mask() {
        unsigned bits = Bits % word_bits;
        if (bits == 0) {
            return ~0ULL;
        }
        return (1ULL << bits) - 1;
    }

    void add_small(unsigned value) {
        unsigned __int128 carry = value;
        for (unsigned i = 0; i < word_count && carry != 0; i++) {
            unsigned __int128 sum = static_cast<unsigned __int128>(words[i]) + carry;
            words[i] = static_cast<unsigned long long>(sum);
            carry = sum >> 64;
        }
        normalize();
    }

    void mul_small(unsigned value) {
        unsigned __int128 carry = 0;
        for (unsigned i = 0; i < word_count; i++) {
            unsigned __int128 product = static_cast<unsigned __int128>(words[i]) * value + carry;
            words[i] = static_cast<unsigned long long>(product);
            carry = product >> 64;
        }
        normalize();
    }

    unsigned div_small(unsigned divisor) {
        unsigned __int128 rem = 0;
        for (unsigned i = word_count; i-- > 0;) {
            unsigned __int128 cur = (rem << 64) | words[i];
            words[i] = static_cast<unsigned long long>(cur / divisor);
            rem = cur % divisor;
        }
        normalize();
        return static_cast<unsigned>(rem);
    }

    static void divmod(const ivy_uint &a, const ivy_uint &b, ivy_uint &q, ivy_uint &r) {
        if (b.is_zero()) {
            throw std::domain_error("ivy_uint division by zero");
        }
        q.clear();
        r.clear();
        for (unsigned i = Bits; i-- > 0;) {
            r <<= 1;
            if (a.bit(i)) {
                r.words[0] |= 1ULL;
            }
            if (r >= b) {
                r -= b;
                q.set_bit(i);
            }
        }
        q.normalize();
        r.normalize();
    }
};

static inline unsigned ivy_bv_shift_amount(unsigned value) {
    return value;
}

static inline unsigned ivy_bv_shift_amount(unsigned long long value) {
    return value > static_cast<unsigned long long>(UINT_MAX) ? UINT_MAX : static_cast<unsigned>(value);
}

#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)
static inline unsigned ivy_bv_shift_amount(unsigned __int128 value) {
    return value > static_cast<unsigned __int128>(UINT_MAX) ? UINT_MAX : static_cast<unsigned>(value);
}
#endif

template <unsigned Bits>
static inline unsigned ivy_bv_shift_amount(const ivy_uint<Bits> &value) {
    return value.to_shift_amount();
}

#endif
