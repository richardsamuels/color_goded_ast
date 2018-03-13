#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

struct Callback {
    void *data_ptr;
    void *err_str;
};

void InsertHighlight(struct Callback c, char * type, int line, int column,
        char * token);

void Errored(struct Callback c, char* msg);

#ifdef __cplusplus
}
#endif
