#include "database.hpp"

string generateUUID() {
    static const char hex[] = "0123456789abcdef";

    static random_device rd;
    static mt19937 gen(rd());
    static uniform_int_distribution<int> hexDist(0, 15);
    static uniform_int_distribution<int> variantDist(8, 11);

    string uuid = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx";

    for (char &c : uuid) {
        if (c == 'x') {
            c = hex[hexDist(gen)];
        } else if (c == 'y') {
            c = hex[variantDist(gen)];
        }
    }

    return uuid;
}

bool likeMatch(const string& str, const string& pattern) {
    size_t i = 0, j = 0;
    size_t star_i = -1, star_j = -1;

    while (i < str.size()) {
        if (j < pattern.size() && (pattern[j] == str[i] || pattern[j] == '_')) {
            i++;
            j++;
        } else if (j < pattern.size() && pattern[j] == '%') {
            star_i = i;
            star_j = j++;
        } else if (star_j != -1) {
            j = star_j + 1;
            i = ++star_i;
        } else {
            return false;
        }
    }

    while (j < pattern.size() && pattern[j] == '%') {
        j++;
    }

    return (j == pattern.size());
}

bool checkCondition(JSONNode& fieldValue, const JSONNode& condition) {
    if (condition.isValue()) {
        // Default $eq
        if (fieldValue.d_type != condition.d_type) return false;
        switch (condition.d_type) {
            case JSONType::NUMBER:
                return fieldValue.get<double>() == condition.get<double>();
            case JSONType::STRING:
                return fieldValue.get<string>() == condition.get<string>();
            case JSONType::BOOL:
                return fieldValue.get<bool>() == condition.get<bool>();
            case JSONType::NULLT:
                return true;
            default: return false;
        }
    }

    if (condition.isObject()) {
        for (auto& [op, val] : condition.d_data.items()) {
            if (op == "$eq") {
                if (!checkCondition(fieldValue, val)) return false;
            } else if (op == "$gt") {
                if (fieldValue.d_type != JSONType::NUMBER || fieldValue.get<double>() <= val.get<double>())
                    return false;
            } else if (op == "$lt") {
                if (fieldValue.d_type != JSONType::NUMBER || fieldValue.get<double>() >= val.get<double>())
                    return false;
            } else if (op == "$like") {
                if (fieldValue.d_type != JSONType::STRING) return false;
                if (!likeMatch(fieldValue.get<string>(), val.get<string>())) return false;
            } else if (op == "$in") {
                if (!val.isArray()) return false;
                bool found = false;
                for (auto& v : val.d_array) {
                    if (checkCondition(fieldValue, v)) {
                        found = true;
                        break;
                    }
                }
                if (!found) return false;
            } else {
                throw runtime_error("Unknown operator: " + op);
            }
        }
        return true;
    }

    return false;
}

bool matchesFilter(JSONNode& doc, const JSONNode& filter) {
    if (!filter.isObject()) return false;

    if (filter.d_data.contains("$or")) {
        const JSONNode& arr = filter.d_data.get("$or");
        for (auto& subFilter : arr.d_array) {
            if (matchesFilter(doc, subFilter)) return true;
        }
        return false;
    }

    if (filter.d_data.contains("$and")) {
        const JSONNode& arr = filter.d_data.get("$and");
        for (auto& subFilter : arr.d_array) {
            if (!matchesFilter(doc, subFilter)) return false;
        }
        return true;
    }

    for (auto& [key, value] : filter.d_data.items()) {
        if (key == "$or" || key == "$and") continue;
        if (!doc.d_data.contains(key) || !checkCondition(doc.d_data.get(key), value))
            return false;
    }

    return true;
}

void handleInsert(JSONNode& document, const string& data) {
    JSONNode newDocument = JSONNode::parse(data);
    if (newDocument.d_type == JSONType::ARRAY) {
        for (auto& elm : newDocument.d_array) {
            string userUUID = "ObjectId(" + generateUUID() + ")";
            elm.d_data.put("_id", static_cast<JSONNode>(userUUID));
            document.d_array.push_back(elm);
        }
    } else {
        string userUUID = "ObjectId(" + generateUUID() + ")";
        newDocument.d_data.put("_id", static_cast<JSONNode>(userUUID));
        document.d_array.push_back(newDocument);
    }
}

void handleDelete(JSONNode& collection, const string& data) {
    JSONNode filter = JSONNode::parse(data);
    JSONNode result(JSONType::ARRAY);

    for (auto &doc : collection.d_array) {
        if (!matchesFilter(doc, filter)) {
            result.appendArray(doc);
        }
    }

    collection = move(result);
}

void handleFind(JSONNode& document, const string& data) {
    JSONNode filter = JSONNode::parse(data);
    JSONNode result(JSONType::ARRAY);

    for (auto& doc : document.d_array) {
        if (matchesFilter(doc, filter)) {
            result.appendArray(doc);
        }
    }

    cout << JSONNode::prettyStringify(result, 0) << endl;
}

void executeCommand(const DBCommand& cmd, JSONNode& document) {
	if (cmd.action == CommandAction::INSERT) {
		handleInsert(document, cmd.json);
	} else if (cmd.action == CommandAction::DELETE) {
		handleDelete(document, cmd.json);
	} else if (cmd.action == CommandAction::FIND) {
		handleFind(document, cmd.json);
	} else {
		throw runtime_error("Unknown action");
	}
}

JSONNode loadCollection(const string& dbName, const string& collectionName) {
    if (!filesystem::exists(dbName)) {
        filesystem::create_directories(dbName);
    }

    string pathToCollection = dbName + "/" + collectionName;

    if (!filesystem::exists(pathToCollection)) {
        ofstream outfile(pathToCollection);
        if (!outfile) {
            throw runtime_error("Cannot create collection file: " + pathToCollection);
        }
        outfile << "[]";
        outfile.close();
    }

    ifstream infile(pathToCollection);
    if (!infile) {
        throw runtime_error("Cannot open collection file: " + pathToCollection);
    }

    string document{istreambuf_iterator<char>(infile), istreambuf_iterator<char>()};
    if (document.empty()) document = "[]";

    return JSONNode::parse(document);
}

void saveCollection(const string& dbName, const string& collectionName, const JSONNode& document) {
    string pathToCollection = dbName + "/" + collectionName;
    ofstream outfile(pathToCollection);
    if (!outfile) {
        throw runtime_error("Cannot open file for writing: " + pathToCollection);
    }

    outfile << JSONNode::prettyStringify(document);
}
