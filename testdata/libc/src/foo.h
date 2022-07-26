#include <stdlib.h>

typedef struct {
    union {
        int __i[14];
        unsigned long __s[7];
    } __u;
} pth_attr_t;

typedef struct pub pub_t;

struct pub {
    int a;
    int b :1,
          :2,
        c :3;
};

typedef size_t foo_t;

foo_t foo();
void unknown();
