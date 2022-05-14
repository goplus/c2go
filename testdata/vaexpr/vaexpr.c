#include <stdio.h>
#include <stdarg.h>

void vprintn2(int n, va_list ap1) {
	int i;
	char *s;
	va_list ap;
	__builtin_va_copy(ap, ap1);
	for (i = 0; i < n; i++) {
		s = va_arg(ap, void*);
		printf("%s\n", s);
	}
	printf("%d, %d, %d\n", va_arg(ap, int), (int)va_arg(ap, unsigned int), va_arg(ap, int));
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
