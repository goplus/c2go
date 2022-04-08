#include <stdio.h>
#include <stdarg.h>

void printn(int n, ...) {
	int i;
    char *s;
	va_list ap;
	va_start(ap, n);
    for (i = 0; i < n; i++) {
        s = va_arg(ap, char*);
        printf("%s\n", s);
    }
	va_end(ap);
}
