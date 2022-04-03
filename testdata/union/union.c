#include <stdio.h>

typedef struct {
    struct {
        int a;
    };
    struct {
        int a;
    } p, q;
} foo;

int main() {
    foo foo;
    foo.a = 1;
    foo.p.a = 2;
    foo.q.a = 3;
    printf(
        "foo.a = %d, foo.p.a = %d, foo.q.a = %d\n",
        foo.a, foo.p.a, foo.q.a);
    return 0;
}
