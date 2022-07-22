#include "foo.h"

struct cookie {
    int a;
};

foo_t foo() {
    struct cookie v = {100};
    return v.a;
}
