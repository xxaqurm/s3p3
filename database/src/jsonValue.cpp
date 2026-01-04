#include "jsonValue.hpp"

JSONNode::Value::operator string() const {
    return d_string;
}
JSONNode::Value::operator double() const {
    return d_number;
}
JSONNode::Value::operator int() const {
    return d_number;
}
JSONNode::Value::operator bool() const {
    return d_bool;
}

void JSONNode::limitToArray() {
    if (!isArray())
        throw runtime_error("this operation is only available to array node");
}

void JSONNode::limitToObject() {
    if (!isObject())
        throw runtime_error("this operation is only available to object node");
}

JSONNode::JSONNode(JSONType type) : d_type(type) {}
JSONNode::JSONNode() : d_type(JSONType::NULLT) {}
JSONNode::JSONNode(nullptr_t value) : JSONNode() {}
JSONNode::JSONNode(double value) : d_type(JSONType::NUMBER) {
    d_value.d_number = value;
}
JSONNode::JSONNode(int value) : d_type(JSONType::NUMBER) {
    d_value.d_number = value;
}
JSONNode::JSONNode(const string& value) : d_type(JSONType::STRING) {
    d_value.d_string = value;
}
JSONNode::JSONNode(const char* value) : d_type(JSONType::STRING) {
    d_value.d_string = value;
}
JSONNode::JSONNode(bool value) : d_type(JSONType::BOOL) {
    d_value.d_bool = value;
}
JSONNode::JSONNode(const vector<JSONNode>& nodes) : d_type(JSONType::ARRAY), d_array(nodes) {}
JSONNode::JSONNode(const JSONNode& node) {
    d_type = node.d_type;
    d_data = node.d_data;
    d_array = node.d_array;
    d_value = node.d_value;
}

JSONNode& JSONNode::operator=(const JSONNode& node) {
    d_type = node.d_type;
    d_data = node.d_data;
    d_array = node.d_array;
    d_value = node.d_value;
    return *this;
}

bool JSONNode::isValue() const {
    return d_type == JSONType::BOOL ||
           d_type == JSONType::NUMBER ||
           d_type == JSONType::STRING ||
           d_type == JSONType::NULLT;
}

bool JSONNode::isObject() const {
    return d_type == JSONType::OBJECT;
}
bool JSONNode::isArray() const {
    return d_type == JSONType::ARRAY;
}
bool JSONNode::isNUll() const {
    return d_type == JSONType::NULLT;
}

void JSONNode::appendArray(const JSONNode &node) {
    d_array.push_back(node);
}

JSONNode& JSONNode::operator[](int index) {
    limitToArray();
    return d_array[index];
}

JSONNode& JSONNode::operator[](const string &key) {
    limitToObject();
    try {
        return d_data.get(key);
    } catch (const runtime_error&) {
        d_data.put(key, JSONNode());
        return d_data.get(key);
    }
}

JSONNode& JSONNode::operator[](const char* key) {
    limitToObject();
    return d_data.get(key);
}

JSONNode::operator string() const {
    return d_value.d_string;
}
JSONNode::operator int() const {
    return d_value.d_number;
}
JSONNode::operator double() const {
    return d_value.d_number;
}
JSONNode::operator bool() const {
    return d_value.d_bool; 
}

size_t JSONNode::size() const {
    if (!isArray())
        throw runtime_error("size() is only available for array");
    return d_array.size();
}

string JSONNode::stringify(const JSONNode& node) {
    switch (node.d_type) {
        case JSONType::BOOL:
            return node.d_value.d_bool ? "true" : "false";
        case JSONType::NULLT:
            return "null";
        case JSONType::NUMBER:
            return to_string(node.d_value.d_number);
        case JSONType::STRING:
            return ("\"" + node.d_value.d_string + "\"");
        case JSONType::ARRAY: {
            if (node.d_array.empty()) {
                return "[]";
            }

            string result = "[";
            for (const auto& value : node.d_array) {
                result += JSONNode::stringify(value);
                result += ",";
            }
            result.back() = ']';
            return result;
        }
        case JSONType::OBJECT: {
            if (node.d_data.empty()) {
                return "{}";
            }

            string result = "{";
            for (const auto [key, value] : node.d_data.items()) {
                result += "\"";
                result += key;
                result += JSONNode::stringify(value);
                result += ",";
            }
            result.back() = '}';
            return result;
        }
        default:
            throw runtime_error("Invalid or unhandled JSONType in stringify");
    }
}

string JSONNode::prettyStringify(const JSONNode &node, int indent) {
    const int indentWidth = 4;
    switch (node.d_type) {
        case JSONType::BOOL:
            return node.d_value.d_bool ? "true" : "false";

        case JSONType::NULLT:
            return "null";

        case JSONType::NUMBER:
            return to_string(node.d_value.d_number);

        case JSONType::STRING:
            return "\"" + escapeString(node.d_value.d_string) + "\"";

        case JSONType::ARRAY: {
            if (node.d_array.empty()) return "[]";

            string res;
            res += "[\n";
            for (size_t i = 0; i < node.d_array.size(); ++i) {
                res += makeIndent(indent + 1, indentWidth);
                res += prettyStringify(node.d_array[i], indent + 1);
                if (i + 1 != node.d_array.size()) res += ",";
                res += "\n";
            }
            res += makeIndent(indent, indentWidth);
            res += "]";
            return res;
        }

        case JSONType::OBJECT: {
            auto items = node.d_data.items();
            if (items.empty()) return "{}";

            string res;
            res += "{\n";
            for (size_t i = 0; i < items.size(); ++i) {
                const auto &kv = items[i];
                res += makeIndent(indent + 1, indentWidth);
                res += "\"";
                res += escapeString(kv.first);
                res += "\": ";
                res += prettyStringify(kv.second, indent + 1);
                if (i + 1 != items.size()) res += ",";
                res += "\n";
            }
            res += makeIndent(indent, indentWidth);
            res += "}";
            return res;
        }
    }

    return "null";
}
