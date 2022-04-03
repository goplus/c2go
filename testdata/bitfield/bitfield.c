#include <stdio.h>

typedef struct {
    unsigned short a :1,
        b :2,
          :3,
        c :5;
    short x :1,
        y :2,
          :3,
        z :5;
} foo;

int main() {
    foo foo;
    foo.a = 1;
    foo.b = 3;
    foo.c = 15;
    foo.x = 1;
    foo.y = 3;
    foo.z = 15;
    printf(
        "foo.a = %d, foo.b = %d, foo.c = %d\n",
        foo.a, foo.b, foo.c);
    printf(
        "foo.x = %d, foo.y = %d, foo.z = %d\n",
        foo.x, foo.y, foo.z);
    return 0;
}
