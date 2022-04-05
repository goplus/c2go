#include <stdio.h>
#include <complex.h>

int main() {
    _Complex double a = 3 + 2*I;
    printf("%f + %fi\n", creal(a), cimag(a));
    return 0;
}
