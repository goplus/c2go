#include <stdio.h>

typedef struct {
    struct {
        int a;
    };
} foo;

int main() {
    foo foo;
    foo.a = 1;
    printf("foo.a = %d\n", foo.a);
    return 0;
}
