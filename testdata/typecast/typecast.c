#include <stdio.h>

int f(int n) {
    return n;
}

void call(void (*f)()) {
    if (f == (void(*)())-1) {
        return;
    }
    printf("n: %d\n", ((int(*)(int))f)(3));
}

static inline int a_ctz_64(unsigned long x)
{
	static const char debruijn64[64] = {
		0, 1, 2, 53, 3, 7, 54, 27, 4, 38, 41, 8, 34, 55, 48, 28,
		62, 5, 39, 46, 44, 42, 22, 9, 24, 35, 59, 56, 49, 18, 29, 11,
		63, 52, 6, 26, 37, 40, 33, 47, 61, 45, 43, 21, 23, 58, 17, 10,
		51, 25, 36, 32, 60, 20, 57, 16, 50, 31, 19, 15, 30, 14, 13, 12
	};
	return debruijn64[(x&-x)*0x022fdd63cc95386dull >> 58];
}

int main() {
    call((void(*)())-1);
    call((void*)f);
    printf("%d\n", a_ctz_64(1353));
    return 0;
}
