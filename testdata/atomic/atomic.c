#include <stdio.h>

int main() {
    int a = 3;
    __atomic_store_n(&a, 100, 0);
    printf("atomic: %d\n", a);
    return 0;
}
