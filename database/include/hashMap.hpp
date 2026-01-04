#pragma once
#include <iostream>
#include <stdexcept>
#include <vector>

using namespace std;

template<typename Key, typename Value>
class HashMap {
private:
    struct Node {
        Key key;
        Value value;
        Node* next;
        Node(const Key& k, const Value& v) : key(k), value(v), next(nullptr) {}
    };

    const double LOAD_FACTOR = 0.75;

    size_t capacity_;
    size_t size_;
    vector<Node*> table;

    size_t hash(const Key& key) const;
    void rehash();

public:
    // Constructors
    HashMap(size_t initialCapacity = 11)
        : capacity_(initialCapacity), size_(0), table(capacity_, nullptr) {}
    HashMap(const HashMap& other);
    ~HashMap();

    // Methods
    void put(const Key& key, const Value& val);
    void remove(const Key& key);
    bool empty() const;
    bool contains(const Key& key) const;
    void clear();

    vector<pair<Key, Value>> items() const;
    Value& get(const Key& key);
    const Value& get(const Key& key) const;

    // Operator
    HashMap<Key, Value>& operator=(const HashMap& other);
};

template<typename Key, typename Value>
bool HashMap<Key, Value>::contains(const Key& key) const {
    size_t idx = hash(key) % capacity_;
    Node* node = table[idx];
    while (node) {
        if (node->key == key) {
            return true;
        }
        node = node->next;
    }
    return false;
}

template<typename Key, typename Value>
HashMap<Key, Value>::HashMap(const HashMap& other)
    : capacity_(other.capacity_), size_(other.size_), table(capacity_, nullptr) {
    for (size_t i = 0; i < capacity_; i++) {
        Node* node = other.table[i];
        Node** last = &table[i];
        while (node) {
            *last = new Node(node->key, node->value);
            last = &((*last)->next);
            node = node->next;
        }
    }
}

template<typename Key, typename Value>
HashMap<Key, Value>::~HashMap() {
    clear();
}

template<typename Key, typename Value>
void HashMap<Key, Value>::clear() {
    for (Node* node : table) {
        while (node) {
            Node* tmp = node;
            node = node->next;
            delete tmp;
        }
    }
    table.assign(capacity_, nullptr);
    size_ = 0;
}

template<typename Key, typename Value>
size_t HashMap<Key, Value>::hash(const Key& key) const {
    if constexpr (is_integral_v<Key>) {
        return (static_cast<size_t>(key) * 37) % capacity_;
    } else if constexpr (is_same_v<Key, string>) {
        size_t hash_ = 0;
        for (char c : key) {
            hash_ = hash_ * 31 + static_cast<size_t>(c);
        }
        return hash_ % capacity_;
    } else {
        throw runtime_error ("Unsupported key type for hash");
    }
}

template<typename Key, typename Value>
void HashMap<Key, Value>::rehash() {
    size_t oldCapacity = capacity_;
    capacity_ = capacity_ * 2 + 1;

    vector<Node*> newTable(capacity_, nullptr);

    for (size_t i = 0; i < oldCapacity; i++) {
        Node* node = table[i];
        while (node) {
            Node* nextNode = node->next;
            size_t idx = hash(node->key) % capacity_;

            node->next = newTable[idx];
            newTable[idx] = node;
            node = nextNode;
        }
    }

    table = move(newTable);
}

template<typename Key, typename Value>
void HashMap<Key, Value>::put(const Key& key, const Value& value) {
    if (static_cast<double>(size_ + 1) / capacity_ >= LOAD_FACTOR) {
        rehash();
    }
    
    size_t idx = hash(key) % capacity_;
    Node* node = table[idx];
    while (node) {
        if (node->key == key) {
            node->value = value;
            return;
        }

        node = node->next;
    }

    Node* newNode = new Node(key, value);
    newNode->next = table[idx];
    table[idx] = newNode;
    size_++;
}

template<typename Key, typename Value>
void HashMap<Key, Value>::remove(const Key& key) {
    size_t idx = hash(key) % capacity_;
    Node* node = table[idx];
    Node* prev = nullptr;
    while (node) {
        if (node->key == key) {
            if (prev) {
                prev->next = node->next;
            } else {
                table[idx] = node->next;
            }
            delete node;
            size_--;
            return;
        }
        prev = node;
        node = node->next;
    }
}

template<typename Key, typename Value>
vector<pair<Key, Value>> HashMap<Key, Value>::items() const {
    vector<pair<Key, Value>> result;
    for (size_t i = 0; i < capacity_; i++) {
        Node* node = table[i];
        while (node) {
            result.push_back({node->key, node->value});
            node = node->next;
        }
    }
    return result;
}

template<typename Key, typename Value>
Value& HashMap<Key, Value>::get(const Key& key) {
    size_t idx = hash(key) % capacity_;
    
    Node* node = table[idx];
    while (node) {
        if (node->key == key) {
            return node->value;
        }
        node = node->next;
    }

    throw runtime_error("Key not found");
}

template<typename Key, typename Value>
const Value& HashMap<Key, Value>::get(const Key& key) const {
    size_t idx = hash(key) % capacity_;
    
    Node* node = table[idx];
    while (node) {
        if (node->key == key) {
            return node->value;
        }
        node = node->next;
    }

    throw runtime_error("Key not found");
}

template<typename Key, typename Value>
bool HashMap<Key, Value>::empty() const {
    Node* node;
    for (size_t i = 0; i < capacity_; i++) {
        node = table[i];
        if (!node) {
            return false;
        }
    }
    return true;
}

template<typename Key, typename Value>
HashMap<Key, Value>& HashMap<Key, Value>::operator=(const HashMap& other) {
    if (this == &other) {
        return *this;
    }
    clear();
    capacity_ = other.capacity_;
    size_ = other.size_;
    table.assign(capacity_, nullptr);

    for (size_t i = 0; i < capacity_; i++) {
        Node* node = other.table[i];
        Node** last = &table[i];
        while (node) {
            *last = new Node(node->key, node->value);
            last = &((*last)->next);
            node = node->next;
        }
    }
    return *this;
}
