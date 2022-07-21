#include <stdio.h>
#include <string.h>

#define likely(x) __builtin_expect(!!(x), 1)
#define unlikely(x) __builtin_expect(!!(x), 0)

void f() {
	double a;
	float b;
	if (a < b) {
		printf("OK!\n");
	}
}

int g(int ret) {
    printf("%d\n", ret);
    return ret;
}

void test() {
    if (likely(g(100))) {
        printf("likely(g(100))\n");
    }
}

void h() {
    g(1) || g(0);
    g(0) || g(1);
    g(1) && g(0);
    g(0) && g(1);
}

#define shunget(f) (rpos ? (void)rpos-- : (void)0)

void unused(int rpos) {
    volatile int y = 0; // same as: (void)y;
    shunget(rpos);
}

typedef struct {
    char msg[10];
} foo_t;

int main() {
    (void)0;
    foo_t foo = {"Hi, c2go!"};
    foo_t *pfoo = &foo;
    char msg[] = {'a', 'b', '\0'};
    char *pmsg = msg;
    printf("%c\n", msg[1]);
    pmsg[1] = (msg[0]>='a'?'!':'?'), printf("%s\n", pmsg),
    pfoo->msg[0] += 'a'-'A',
    pfoo->msg[2] = '!', printf("%s\n", foo.msg);
    h();
    printf("%d\n", (int)((-0x2000ULL << (8*sizeof(long)-1)) | (4096ULL -1)));
    test();
    return 0;
}
