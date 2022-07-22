#include "foo.h"

struct cookie {
    double a;
    int b;
};

static foo_t foo2() {
    struct cookie v = {0, 10};
    return v.b;
}
