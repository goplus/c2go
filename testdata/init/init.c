#include <stdio.h>

static const int k = 1023;

int g_a = 100;

static int g_arr[][3];
static int g_arr[100][3];

int (*g_parr)[100][3] = &g_arr;

union { int a; unsigned short b; } g_foo = { 1 };
union { short a; char b[10]; } g_bar = { .b = "Hello" };

struct { int a; unsigned short b; } g_x = { 2, 3 };
struct { short a; char b[10]; int c; } g_y = { .b = "Hi", .a = 11 };

struct { int a[3], b; } v[] = { [0].a = {1}, [1].a[0] = 2 };
struct { int a[3], b; } w[] = { {{1000, 0, 0}}, {{2000, 0, 0}} };

struct test_t { int a[3], b; } test = { .b = 1234 };

int main() {
    int a = 100;

    union { int a; unsigned short b; } foo = { 1 };
    union { short a; char b[10]; } bar = { .b = "Hello" };

    struct { int a; unsigned short b; } x = { 2, 3 };
    struct { short a; int b; char c[10]; } y = { .c = "Hi", .a = 11 };

    struct test_t z = test;

    printf("a = %d, g_a = %d, k = %d\n", a, g_a, k);

    printf("foo.a = %d, foo.b = %d\n", foo.a, foo.b);
    printf("bar.a = %d, bar.b = %s\n", bar.a, bar.b);

    printf("g_foo.a = %d, g_foo.b = %d\n", g_foo.a, g_foo.b);
    printf("g_bar.a = %d, g_bar.b = %s\n", g_bar.a, g_bar.b);

    printf("x.a = %d, x.b = %d\n", x.a, x.b);
    printf("y.a = %d, y.c = %s\n", y.a, y.c);

    printf("g_x.a = %d, g_x.b = %d\n", g_x.a, g_x.b);
    printf("g_y.a = %d, g_y.b = %s\n", g_y.a, g_y.b);

    printf("w[0].a[0] = %d, w[1].a[0] = %d\n", w[0].a[0], w[1].a[0]);
    printf("z.b = %d\n", z.b);
    return 0;
}
