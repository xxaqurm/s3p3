#pragma once
#include <iostream>
#include <vector>
#include <string>
#include <cstdint>

#include "hashMap.hpp"
#include "utils.hpp"

using namespace std;

enum class JSONType : uint8_t {
    NUMBER,
    STRING,
    NULLT,
    OBJECT,
    ARRAY,
    BOOL
};

class JSONNode {
private:
    struct Value {
        string d_string{};
        double d_number{0.0};
        bool d_bool{false};

        operator string() const;
        operator double() const;
        operator int() const;
        operator bool() const;
    } d_value;

    void limitToArray();
    void limitToObject();

public:
    JSONType d_type;
    HashMap<string, JSONNode> d_data;
    vector<JSONNode> d_array;

    // Constructors
    JSONNode(JSONType type);
    JSONNode();
    JSONNode(nullptr_t value);
    JSONNode(double value);
    JSONNode(int value);
    JSONNode(const string& value);
    JSONNode(const char* value);
    JSONNode(bool value);
    JSONNode(const vector<JSONNode>& nodes);
    JSONNode(const JSONNode& node);

    // Methods
    bool isValue() const;
    bool isObject() const;
    bool isArray() const;
    bool isNUll() const;
    void appendArray(const JSONNode &node);

    template<typename T>
    T get() {
        if (!isValue()) {
            throw std::runtime_error("unable to get value for this type");
        }
        return static_cast<T>(d_value);
    }

    template<typename T>
    T get() const {
        if (!isValue()) {
            throw std::runtime_error("unable to get value for this type");
        }
        return static_cast<T>(d_value);
    }

    size_t size() const;

    static JSONNode parse(const string& s);
    static string stringify(const JSONNode& node);
    static string prettyStringify(const JSONNode& node, int indent = 0);

    // Operators
    JSONNode& operator=(const JSONNode& node);
    JSONNode& operator[](int index);
    JSONNode& operator[](const string& key);
    JSONNode& operator[](const char* key);

    operator string() const;
    operator int() const;
    operator double() const;
    operator bool() const;
};
