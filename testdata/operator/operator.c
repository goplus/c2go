#include <stdio.h>
#include <string.h>

/* shl() need n > 0 */
static inline void shl(size_t p[2], int n) {
	if(n >= 8 * sizeof(size_t)) {
		n -= 8 * sizeof(size_t);
		p[1] = p[0];
		p[0] = 0;
	}
	p[1] <<= n;
	p[1] |= p[0] >> (sizeof(size_t) * 8 - n);
	p[0] <<= n;
}

static inline int a_clz_64(unsigned long long x) {
	unsigned int y;
	int r;
	if (x>>32) y=x>>32, r=0; else y=x, r=32;
	if (y>>16) y>>=16; else r |= 16;
	if (y>>8) y>>=8; else r |= 8;
	if (y>>4) y>>=4; else r |= 4;
	if (y>>2) y>>=2; else r |= 2;
	return r | !(y>>1);
}

static void cycle(size_t width, unsigned char* ar[], int n) {
	unsigned char tmp[256];
	size_t l;
	int i;

	if(n < 2) {
		return;
	}

	ar[n] = tmp;
	while(width) {
		l = sizeof(tmp) < width ? sizeof(tmp) : width;
		memcpy(ar[n], ar[0], l);
		for(i = 0; i < n; i++) {
			memcpy(ar[i], ar[i + 1], l);
			ar[i] += l;
		}
		width -= l;
	}
}

typedef struct {
    char msg[10];
} foo_t;

int main() {
	(void)0;
    foo_t foo = {"Hi, c2go!"};
    foo_t *pfoo = &foo;
    char msg[] = {'a', 'b', '\0'};
    char *pmsg = msg;
    printf("%c\n", msg[1]);
    pmsg[1] = (msg[0]>='a'?'!':'?'), printf("%s\n", pmsg),
    pfoo->msg[0] += 'a'-'A',
    pfoo->msg[2] = '!', printf("%s\n", foo.msg);
    return 0;
}
