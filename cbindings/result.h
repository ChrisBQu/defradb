// result.h
#ifndef RESULT_H
#define RESULT_H

typedef struct {
    int status;
    char* error;
    char* value;
} Result;

typedef struct {
    unsigned long long tx;
    const char* version;
    const char* schema;
    const char* name;
    const char* identity;
    int getInactive;
} CollectionOptions;

typedef struct {
    const char* dbPath;
    const char* listeningAddresses;
} NodeInitOptions;

#endif // RESULT_H
