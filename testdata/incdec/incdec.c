#include <stdio.h>

int main() {
    int a = 1;
    int *pa = &a;
    (*pa)++;
    printf("%d\n", a);
    return 0;
}
