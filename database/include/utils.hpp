#pragma once
#include <string>
#include <vector>
#include <cctype>

#include "jsonValue.hpp"
#include "hashMap.hpp"

using namespace std;

void findBracePairs(const string& s, HashMap<int, int>& bracePairs);
bool isWhitespace(char c);
bool isDouble(const string& s);
bool isInteger(const string& s);

string escapeString(const string &s);
string makeIndent(int level, int indentWidth = 4);
