#include <stdio.h>

int g_a = 100;

int g_arr[100][3];
int (*g_parr)[100][3] = &g_arr;

union { int a; unsigned short b; } g_foo = { 1 };
union { short a; char b[10]; } g_bar = { .b = "Hello" };

struct { int a; unsigned short b; } g_x = { 2, 3 };
struct { short a; char b[10]; int c; } g_y = { .b = "Hi", .a = 11 };

int main() {
    int a = 100;

    union { int a; unsigned short b; } foo = { 1 };
    union { short a; char b[10]; } bar = { .b = "Hello" };

    struct { int a; unsigned short b; } x = { 2, 3 };
    struct { short a; int b; char c[10]; } y = { .c = "Hi", .a = 11 };

    printf("a = %d, g_a = %d\n", a, g_a);

    printf("foo.a = %d, foo.b = %d\n", foo.a, foo.b);
    printf("bar.a = %d, bar.b = %s\n", bar.a, bar.b);

    printf("g_foo.a = %d, g_foo.b = %d\n", g_foo.a, g_foo.b);
    printf("g_bar.a = %d, g_bar.b = %s\n", g_bar.a, g_bar.b);

    printf("x.a = %d, x.b = %d\n", x.a, x.b);
    printf("y.a = %d, y.b = %s\n", y.a, y.b);

    printf("g_x.a = %d, g_x.b = %d\n", g_x.a, g_x.b);
    printf("g_y.a = %d, g_y.b = %s\n", g_y.a, g_y.b);
    return 0;
}
