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

struct s_file {
    struct s_file_methods *vtable;
};

struct s_file_methods {
    int a;
};

struct s_file_methods;

int main() {
    struct {
        int a;
    } e, f;
    foo foo;
    foo.a = 1;
    foo.p.a = 2;
    foo.y.a = 3;
    e.a = 100;
    f.a = 101;
    printf(
        "foo.a = %d, foo.p.a = %d, foo.y.a = %d\n",
        foo.a, foo.p.a, foo.y.a);
    printf(
        "e.a = %d, f.a = %d\n",
        e.a, f.a);
    return 0;
}
