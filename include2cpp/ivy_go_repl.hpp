#pragma once

#include <sstream>
#include <stdexcept>
#include <string>
#include <vector>

static bool ivy2cpp_is_white(char c) {
    return c == ' ' || c == '\t' || c == '\n' || c == '\r';
}

static bool ivy2cpp_is_ident(char c) {
    return c == '_' || c == '.' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9');
}

static void ivy2cpp_skip_white(const std::string &s, size_t &pos) {
    while (pos < s.size() && ivy2cpp_is_white(s[pos])) {
        pos++;
    }
}

static std::string ivy2cpp_trim_arg(const std::string &s) {
    size_t begin = 0;
    while (begin < s.size() && ivy2cpp_is_white(s[begin])) {
        begin++;
    }
    size_t end = s.size();
    while (end > begin && ivy2cpp_is_white(s[end - 1])) {
        end--;
    }
    return s.substr(begin, end - begin);
}

static std::string ivy2cpp_read_ident(const std::string &s, size_t &pos) {
    std::string res;
    while (pos < s.size() && ivy2cpp_is_ident(s[pos])) {
        res.push_back(s[pos++]);
    }
    if (res.empty()) {
        throw std::runtime_error("expected identifier");
    }
    return res;
}

static std::string ivy2cpp_read_value(const std::string &line, size_t &pos) {
    ivy2cpp_skip_white(line, pos);
    size_t begin = pos;
    int depth = 0;
    bool in_string = false;
    bool escape = false;
    for (; pos < line.size(); pos++) {
        char c = line[pos];
        if (in_string) {
            if (escape) {
                escape = false;
            } else if (c == '\\') {
                escape = true;
            } else if (c == '"') {
                in_string = false;
            }
            continue;
        }
        if (c == '"') {
            in_string = true;
        } else if (c == '[' || c == '{') {
            depth++;
        } else if (c == ']' || c == '}') {
            if (depth > 0) {
                depth--;
            }
        } else if (depth == 0 && (c == ',' || c == ')')) {
            break;
        }
    }
    std::string value = ivy2cpp_trim_arg(line.substr(begin, pos - begin));
    if (value.empty()) {
        throw std::runtime_error("missing argument");
    }
    return value;
}

static void ivy2cpp_parse_command(const std::string &line, std::string &action, std::vector<std::string> &args) {
    args.clear();
    size_t pos = 0;
    ivy2cpp_skip_white(line, pos);
    action = ivy2cpp_read_ident(line, pos);
    ivy2cpp_skip_white(line, pos);
    if (pos == line.size()) {
        return;
    }
    if (line[pos] != '(') {
        throw std::runtime_error("expected '(' after action");
    }
    pos++;
    while (true) {
        args.push_back(ivy2cpp_read_value(line, pos));
        ivy2cpp_skip_white(line, pos);
        if (pos >= line.size()) {
            throw std::runtime_error("expected ')'");
        }
        if (line[pos] == ')') {
            pos++;
            break;
        }
        if (line[pos] != ',') {
            throw std::runtime_error("expected ','");
        }
        pos++;
    }
    ivy2cpp_skip_white(line, pos);
    if (pos != line.size()) {
        throw std::runtime_error("trailing text after command");
    }
}

static std::string ivy2cpp_read_arg(const std::vector<std::string> &args, size_t idx, const char *name) {
    if (idx >= args.size()) {
        throw std::runtime_error(std::string("missing argument: ") + name);
    }
    return args[idx];
}

static void ivy2cpp_check_arity(const std::vector<std::string> &args, size_t expected, const std::string &action) {
    if (args.size() != expected) {
        std::ostringstream msg;
        msg << "action " << action << " takes " << expected << " input parameters";
        throw std::runtime_error(msg.str());
    }
}
