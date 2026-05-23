#pragma once

#include <cstddef>
#include <fstream>
#include <functional>
#include <map>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <utility>
#include <vector>

namespace hash_space {

inline unsigned string_hash(const char *str, unsigned length, unsigned init_value) {
    unsigned h = init_value ? init_value : 2166136261u;
    for (unsigned i = 0; i < length; i++) {
        h ^= static_cast<unsigned char>(str[i]);
        h *= 16777619u;
    }
    return h;
}

template <typename T> class hash {
public:
    size_t operator()(const T &s) const {
        return s.__hash();
    }
};

template <> class hash<int> {
public:
    size_t operator()(const int &s) const { return static_cast<size_t>(s); }
};

template <> class hash<long long> {
public:
    size_t operator()(const long long &s) const { return static_cast<size_t>(s); }
};

template <> class hash<unsigned> {
public:
    size_t operator()(const unsigned &s) const { return static_cast<size_t>(s); }
};

template <> class hash<unsigned long long> {
public:
    size_t operator()(const unsigned long long &s) const { return static_cast<size_t>(s); }
};

#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)
template <> class hash<unsigned __int128> {
public:
    size_t operator()(const unsigned __int128 &s) const {
        unsigned long long lo = static_cast<unsigned long long>(s);
        unsigned long long hi = static_cast<unsigned long long>(s >> 64);
        return static_cast<size_t>(lo ^ (hi + 0x9e3779b97f4a7c15ULL + (lo << 6) + (lo >> 2)));
    }
};
#endif

template <> class hash<bool> {
public:
    size_t operator()(const bool &s) const { return static_cast<size_t>(s); }
};

template <> class hash<std::string> {
public:
    size_t operator()(const std::string &s) const {
        return string_hash(s.c_str(), static_cast<unsigned>(s.size()), 0);
    }
};

template <> class hash<std::pair<int, int> > {
public:
    size_t operator()(const std::pair<int, int> &p) const {
        return static_cast<size_t>(p.first + p.second);
    }
};

template <typename T> class hash<std::vector<T> > {
public:
    size_t operator()(const std::vector<T> &p) const {
        hash<T> h;
        size_t res = 0;
        for (typename std::vector<T>::const_iterator it = p.begin(), en = p.end(); it != en; ++it) {
            res += h(*it);
        }
        return res;
    }
};

template <typename K, typename V> class hash<std::map<K, V> > {
public:
    size_t operator()(const std::map<K, V> &p) const {
        hash<K> hk;
        hash<V> hv;
        size_t res = 0;
        for (typename std::map<K, V>::const_iterator it = p.begin(), en = p.end(); it != en; ++it) {
            res += hk(it->first) + hv(it->second);
        }
        return res;
    }
};

template <class T> class hash<std::pair<T *, T *> > {
public:
    size_t operator()(const std::pair<T *, T *> &p) const {
        return reinterpret_cast<size_t>(p.first) + reinterpret_cast<size_t>(p.second);
    }
};

template <class T> class hash<T *> {
public:
    size_t operator()(T *const &p) const {
        return reinterpret_cast<size_t>(p);
    }
};

template <typename T> class equal {
public:
    bool operator()(const T &x, const T &y) const { return x == y; }
};

template <typename Element, class HashFun = hash<Element>, class EqFun = equal<Element> >
class hash_set : public std::unordered_set<Element, HashFun, EqFun> {
public:
    typedef std::unordered_set<Element, HashFun, EqFun> base_type;
    typedef Element value_type;

    hash_set() : base_type() {}
};

template <typename Key, typename Value, class HashFun = hash<Key>, class EqFun = equal<Key> >
class hash_map : public std::unordered_map<Key, Value, HashFun, EqFun> {
public:
    typedef std::unordered_map<Key, Value, HashFun, EqFun> base_type;

    hash_map() : base_type() {}
};

template <typename D, typename R, class HashFun, class EqFun>
class hash<hash_map<D, R, HashFun, EqFun> > {
public:
    size_t operator()(const hash_map<D, R, HashFun, EqFun> &p) const {
        hash<D> h1;
        hash<R> h2;
        size_t res = 0;
        for (typename hash_map<D, R, HashFun, EqFun>::const_iterator it = p.begin(), en = p.end(); it != en; ++it) {
            res += h1(it->first) + h2(it->second);
        }
        return res;
    }
};

template <typename D, typename R, class HashFun, class EqFun>
inline bool operator==(const hash_map<D, R, HashFun, EqFun> &s, const hash_map<D, R, HashFun, EqFun> &t) {
    const typename hash_map<D, R, HashFun, EqFun>::base_type &sb = s;
    const typename hash_map<D, R, HashFun, EqFun>::base_type &tb = t;
    return sb == tb;
}

} // namespace hash_space
