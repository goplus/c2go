#include <stdio.h>

int main() {
    long long a = 3;
    int b = 0;
    __atomic_store_n(&a, 100, 0);
    printf("atomic: %lld\n", a);
    __atomic_store_n(&b, a!=0, 0);
    printf("atomic: %d\n", __atomic_load_n(&b, 0));
    return 0;
}
