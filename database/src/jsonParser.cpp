#include "jsonParser.hpp"

void skipWS(const string& s, size_t& i) {
    while (i < s.size() && isWhitespace(s[i])) i++;
}

string parseString(const string& s, size_t& i) {
    string out;

    i++;
    while (i < s.size() && s[i] != '"') {
        out.push_back(s[i]);
        i++;
    }
    i++;
    return out;
}

double parseNumber(const string& s, size_t& i) {
    size_t start = i;

    if (i < s.size() && (s[i] == '-' || s[i] == '+')) i++;
    while (i < s.size() && isdigit(s[i])) i++;
    if (i < s.size() && s[i] == '.') {
        i++;
        while (i < s.size() && isdigit(s[i])) i++;
    }

    if (i < s.size() && (s[i] == 'e' || s[i] == 'E')) {
        i++;
        if (i < s.size() && (s[i] == '+' || s[i] == '-')) i++;
        while (i < s.size() && isdigit(s[i])) i++;
    }

    return stod(s.substr(start, i - start));
}

JSONNode parseValue(const string& s, size_t& i) {
    skipWS(s, i);

    if (i < s.size() && s[i] == '"')
        return JSONNode(parseString(s, i));

    if (i < s.size() && s[i] == '{')
        return parseObject(s, i);

    if (i < s.size() && s[i] == '[')
        return parseArray(s, i);

    if (i + 3 < s.size() && s.compare(i, 4, "null") == 0) {
        i += 4;
        return JSONNode(nullptr);
    }

    if (i + 3 < s.size() && s.compare(i, 4, "true") == 0) {
        i += 4;
        return JSONNode(true);
    }

    if (i + 4 < s.size() && s.compare(i, 5, "false") == 0) {
        i += 5;
        return JSONNode(false);
    }

    return JSONNode(parseNumber(s, i));
}

JSONNode parseObject(const string& s, size_t& i) {
    JSONNode obj(JSONType::OBJECT);

    i++; // skip {
    skipWS(s, i);

    if (s[i] == '}') {
        i++;
        return obj;
    }

    while (i < s.size()) {
        skipWS(s, i);
        string key = parseString(s, i);

        skipWS(s, i);
        if (s[i] != ':')
            throw runtime_error("Expected ':' in object");
        i++;

        JSONNode value = parseValue(s, i);
        obj[key] = value;

        skipWS(s, i);
        if (s[i] == '}') {
            i++;
            break;
        }
        if (s[i] != ',')
            throw runtime_error("Expected ',' in object");
        i++; // skip comma
    }

    return obj;
}

JSONNode parseArray(const string& s, size_t& i) {
    JSONNode arr(JSONType::ARRAY);

    i++; // skip [
    skipWS(s, i);

    if (s[i] == ']') { i++; return arr; }

    while (i < s.size()) {
        arr.appendArray(parseValue(s, i));

        skipWS(s, i);
        if (s[i] == ']') {
            i++;
            break;
        }
        if (s[i] != ',')
            throw runtime_error("Expected ',' in array");
        i++;
    }

    return arr;
}

JSONNode JSONNode::parse(const string& s) {
    size_t i = 0;
    skipWS(s, i);
    return parseValue(s, i);
}
