#pragma once
#include <fstream>
#include <sstream>
#include <random>
#include <iomanip>
#include <filesystem>

#include "jsonParser.hpp"
#include "jsonValue.hpp"
#include "queryParser.hpp"

#include "utils.hpp"

using namespace std;

JSONNode loadCollection(const string& dbName, const string& collectionName);
void saveCollection(const string& dbName, const string& collectionName, const JSONNode& document);
void executeCommand(const DBCommand& cmd, JSONNode& data);
