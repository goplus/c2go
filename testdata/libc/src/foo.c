#include "foo.h"

struct cookie {
    int a;
};

foo_t foo() {
    pth_attr_t pth;
    pth.__u.__i[0] = 100;
    struct cookie v = {pth.__u.__i[0]};
    return v.a;
}
