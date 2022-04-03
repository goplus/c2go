#include <stdio.h>

typedef struct {
    struct {
        int a;
    };
    struct {
        int a;
    } p, q;
    struct bar {
        int a;
    } x, y;
} foo;

int main() {
    foo foo;
    foo.a = 1;
    foo.p.a = 2;
    foo.y.a = 3;
    printf(
        "foo.a = %d, foo.p.a = %d, foo.y.a = %d\n",
        foo.a, foo.p.a, foo.y.a);
    return 0;
}
