#pragma once

#include "ivy_threads.hpp"
#include "ivy_value.hpp"

#include <fstream>
#include <iostream>
#include <string>
#include <vector>

#ifdef _WIN32
#include <io.h>
#else
#include <unistd.h>
#endif

extern std::ofstream __ivy_out;

inline int ask_ret(long long bound) {
    int res;
    while (true) {
        __ivy_out << "? ";
        std::cin >> res;
        if (res >= 0 && res < bound)
            return res;
        std::cerr << "value out of range" << std::endl;
    }
}

inline bool is_white(int c) {
    return (c == ' ' || c == '\t' || c == '\n' || c == '\r');
}

inline bool is_ident(int c) {
    return c == '_' || c == '.' || (c >= 'A' && c <= 'Z')
        || (c >= 'a' && c <= 'z')
        || (c >= '0' && c <= '9');
}

inline void skip_white(const std::string &str, int &pos) {
    while (pos < static_cast<int>(str.size()) && is_white(str[pos]))
        pos++;
}

struct syntax_error {
    int pos;
    syntax_error(int pos) : pos(pos) {}
};

inline void throw_syntax(int pos) {
    throw syntax_error(pos);
}

inline std::string get_ident(const std::string &str, int &pos) {
    std::string res = "";
    while (pos < static_cast<int>(str.size()) && is_ident(str[pos])) {
        res.push_back(str[pos]);
        pos++;
    }
    if (res.size() == 0)
        throw_syntax(pos);
    return res;
}

inline ivy_value parse_value(const std::string &cmd, int &pos) {
    ivy_value res;
    res.pos = pos;
    skip_white(cmd, pos);
    if (pos < static_cast<int>(cmd.size()) && cmd[pos] == '[') {
        while (true) {
            pos++;
            skip_white(cmd, pos);
            if (pos < static_cast<int>(cmd.size()) && cmd[pos] == ']')
                break;
            res.fields.push_back(parse_value(cmd, pos));
            skip_white(cmd, pos);
            if (pos < static_cast<int>(cmd.size()) && cmd[pos] == ']')
                break;
            if (!(pos < static_cast<int>(cmd.size()) && cmd[pos] == ','))
                throw_syntax(pos);
        }
        pos++;
    }
    else if (pos < static_cast<int>(cmd.size()) && cmd[pos] == '{') {
        while (true) {
            ivy_value field;
            pos++;
            skip_white(cmd, pos);
            field.atom = get_ident(cmd, pos);
            skip_white(cmd, pos);
            if (!(pos < static_cast<int>(cmd.size()) && cmd[pos] == ':'))
                 throw_syntax(pos);
            pos++;
            skip_white(cmd, pos);
            field.fields.push_back(parse_value(cmd, pos));
            res.fields.push_back(field);
            skip_white(cmd, pos);
            if (pos < static_cast<int>(cmd.size()) && cmd[pos] == '}')
                break;
            if (!(pos < static_cast<int>(cmd.size()) && cmd[pos] == ','))
                throw_syntax(pos);
        }
        pos++;
    }
    else if (pos < static_cast<int>(cmd.size()) && cmd[pos] == '"') {
        pos++;
        res.atom = "";
        while (pos < static_cast<int>(cmd.size()) && cmd[pos] != '"') {
            char c = cmd[pos++];
            if (c == '\\') {
                if (pos == static_cast<int>(cmd.size()))
                    throw_syntax(pos);
                c = cmd[pos++];
                c = (c == 'n') ? 10 : (c == 'r') ? 13 : (c == 't') ? 9 : c;
            }
            res.atom.push_back(c);
        }
        if (pos == static_cast<int>(cmd.size()))
            throw_syntax(pos);
        pos++;
    }
    else
        res.atom = get_ident(cmd, pos);
    return res;
}

inline void parse_command(const std::string &cmd, std::string &action, std::vector<ivy_value> &args) {
    int pos = 0;
    skip_white(cmd, pos);
    action = get_ident(cmd, pos);
    skip_white(cmd, pos);
    if (pos < static_cast<int>(cmd.size()) && cmd[pos] == '(') {
        pos++;
        skip_white(cmd, pos);
        args.push_back(parse_value(cmd, pos));
        while (true) {
            skip_white(cmd, pos);
            if (!(pos < static_cast<int>(cmd.size()) && cmd[pos] == ','))
                break;
            pos++;
            args.push_back(parse_value(cmd, pos));
        }
        if (!(pos < static_cast<int>(cmd.size()) && cmd[pos] == ')'))
            throw_syntax(pos);
        pos++;
    }
    skip_white(cmd, pos);
    if (pos != static_cast<int>(cmd.size()))
        throw_syntax(pos);
}

struct bad_arity {
    std::string action;
    int num;
    bad_arity(std::string &_action, unsigned _num) : action(_action), num(_num) {}
};

inline void check_arity(std::vector<ivy_value> &args, unsigned num, std::string &action) {
    if (args.size() != num)
        throw bad_arity(action, num);
}

class stdin_reader: public reader {
    std::string buf;
    std::string eof_flag;

public:
    bool eof() {
      return eof_flag.size();
    }
    virtual int fdes() {
        return 0;
    }
    virtual void read() {
        char tmp[257];
        int chars = ::read(0, tmp, 256);
        if (chars == 0) {
            if (buf.size())
                process(buf);
            eof_flag = "eof";
        }
        tmp[chars] = 0;
        buf += std::string(tmp);
        size_t pos;
        while ((pos = buf.find('\n')) != std::string::npos) {
            std::string line = buf.substr(0, pos + 1);
            buf.erase(0, pos + 1);
            process(line);
        }
    }
    virtual void process(const std::string &line) {
        __ivy_out << line;
    }
};
