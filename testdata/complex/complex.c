#include <stdio.h>
#include <complex.h>

int main() {
#ifdef _MSC_VER
    printf("skip complex for msvc\n");
#else
    _Complex double a = 3 + 2*I;
    printf("%f + %fi\n", creal(a), cimag(a));
#endif
    return 0;
}
