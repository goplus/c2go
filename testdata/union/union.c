#include <stdio.h>

typedef union {
    int a;
    unsigned short b;
    struct bar {
        unsigned short low, high;
    } c;
    struct {
        unsigned short low;
        struct {
            unsigned short high;
        };
    };
} foo;

int main() {
    foo foo;
    foo.a = 1;
    foo.b = 2;
    foo.c.high = 3;
    printf(
        "foo.a = %d, foo.b = %d, foo.c.low = %d, foo.c.high = %d\n",
        foo.a, foo.b, foo.c.low, foo.c.high);
    printf(
        "foo.low = %d, foo.high = %d\n",
        foo.low, foo.high);
    foo.low = 5;
    foo.high = 6;
    printf(
        "foo.a = %d, foo.b = %d, foo.c.low = %d, foo.c.high = %d\n",
        foo.a, foo.b, foo.c.low, foo.c.high);
    printf(
        "foo.low = %d, foo.high = %d\n",
        foo.low, foo.high);
    return 0;
}
