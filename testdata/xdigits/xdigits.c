#include <stdio.h>

static const char xdigits[16] = "0123456789ABCDEF";

static const char ydigits[16] = {
    "0123456789ABCDEF"
};

void f(const char* digits) {
    int i;
    for (i = 0; i < 16; i++) {
        printf("%c", digits[i]);
    }
    printf("\n");
}

int main() {
    f(xdigits);
    f(ydigits);
    return 0;
}
