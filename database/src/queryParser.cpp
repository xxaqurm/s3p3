#include "queryParser.hpp"

DBCommand parseQuery(int argc, char* argv[]) {
    if (argc < 4) {
        throw "Usage: ./no_sql_dbms <db> <collection> <command> [json]";
    }

    DBCommand cmd;
    cmd.database = argv[1];
    cmd.collection = argv[2];
    string actionStr = argv[3];
    cmd.json = argc > 4 ? argv[4] : "{}";

    if (actionStr == "insert") {
        cmd.action = CommandAction::INSERT;
    } else if (actionStr == "find") {
        cmd.action = CommandAction::FIND;
    } else if (actionStr == "delete") {
        cmd.action = CommandAction::DELETE;
    } else {
        cmd.action = CommandAction::UNKNOWN;
    }

    return cmd;
}