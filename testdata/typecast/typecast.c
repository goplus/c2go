#include <stdio.h>

int f(int n) {
    return n;
}

void call(void (*f)()) {
    printf("n: %d\n", ((int(*)(int))f)(3));
}

int main() {
    call((void*)f);
    return 0;
}
