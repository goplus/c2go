#include <stdio.h>
#include <stdarg.h>

void vprintn2(int n, va_list ap) {
	int i;
	char *s;
	for (i = 0; i < n; i++) {
		s = va_arg(ap, char*);
		printf("%s\n", s);
	}
}

void vprintn1(int n, va_list ap) {
	vprintn2(n, ap);
}

void printn(int n, ...) {
	va_list ap;
	va_start(ap, n);
	vprintn1(n, ap);
	va_end(ap);
}
