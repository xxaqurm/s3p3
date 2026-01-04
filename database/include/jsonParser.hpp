#pragma once
#include "jsonValue.hpp"
#include "utils.hpp"

using namespace std;

JSONNode parseValue(const string& s, size_t& i);
JSONNode parseObject(const string& s, size_t& i);
JSONNode parseArray(const string& s, size_t& i);
string parseString(const string& s, size_t& i);
double parseNumber(const string& s, size_t& i);
