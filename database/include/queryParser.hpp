#pragma once
#include <iostream>
#include <stdexcept>

using namespace std;

enum class CommandAction {
    INSERT,
    FIND,
    DELETE,
    UNKNOWN
};

struct DBCommand {
    string database;
    string collection;
    CommandAction action;
    string json;
};

DBCommand parseQuery(int argc, char* argv[]);
