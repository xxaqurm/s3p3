#include "utils.hpp"

void findBracePairs(const string &s, HashMap<int, int> &bracePairs) {
    vector<int> stack;
    int n = s.length();

    for (int i = 0; i < n; i++) {
        if (s[i] == '[' || s[i] == '{') {
            stack.push_back(i);
        } else if (s[i] == ']' || s[i] == '}') {
            bracePairs.get(stack.back()) = i;
            stack.pop_back();
        }
    }
}

bool isWhitespace(char c) {
    return c == ' ' || c == '\n' || c == '\t' || c == '\r';
}

bool isDouble(const string& s) {
    size_t i = 0;
    
    if (s.empty()) {
        return false;
    }
    if (s[0] == '+' || s[0] == '-') {
        i++;
    }

    bool dotExist = false;
    while (i < s.length()) {
        if (!isdigit(s[i]) && s[i] != '.') {
            return false;
        }

        if (s[i] == '.') {
            if (dotExist) {
                return false;
            }
            dotExist = true;
        }
        i++;
    }

    return true;
}

bool isInteger(const string& s) {
    size_t i = 0;

    if (s.empty()) {
        return false;
    }
    if (s[0] == '+' || s[0] == '-') {
        i++;
    }

    while (i < s.length()) {
        if (!isdigit(s[i])) {
            return false;
        }
        i++;
    }

    return true;
}

JSONNode getValue(const string& s) {
    int i = 0;
    int j = s.length() - 1;

    while (i <= j && isWhitespace(s[i])) {  // left border
        i++;
    }
    while (j >= i && isWhitespace(s[j])) {  // right border
        j--;
    }

    string value = s.substr(i, j - i + 1);

    if (value[0] == '"') {
        return static_cast<JSONNode>(value.substr(1, value.length() - 2));
    }
    if (value == "true" || value == "false") {
        return static_cast<JSONNode>(value == "true");
    }
    if (value == "null") {
        return JSONNode();
    }

    if (isDouble(value)) {
        return static_cast<JSONNode>(stod(value));
    }
    if (isInteger(value)) {
        return static_cast<JSONNode>(stoi(value));
    }

    throw invalid_argument("Invalid JSON primitive value: " + value);
}

string escapeString(const string& s) {
    // Recording of non-displayed characters
    string out;
    for (unsigned char c : s) {
        switch (c) {
            case '\"': out += "\\\""; break;
            case '\\': out += "\\\\"; break;
            case '\b': out += "\\b";  break;
            case '\f': out += "\\f";  break;
            case '\n': out += "\\n";  break;
            case '\r': out += "\\r";  break;
            case '\t': out += "\\t";  break;
            default:   out += c;      break;
        }
    }
    return out;
}

string makeIndent(int level, int indentWidth) {
    // Make indents
    int total = level * indentWidth;
    if (total <= 0) {
        return string();
    }
    return string(static_cast<size_t>(total), ' ');
}
