#include <stdarg.h>

typedef signed long long intmax_t;
typedef unsigned long long uintmax_t;
typedef unsigned long uintptr_t;
typedef unsigned long size_t;
typedef int ptrdiff_t;

union arg
{
 uintmax_t i;
 long double f;
 void *p;
};

enum {
 BARE, LPRE, LLPRE, HPRE, HHPRE, BIGLPRE,
 ZTPRE, JPRE,
 STOP,
 PTR, INT, UINT, ULLONG,
 LONG, ULONG,
 SHORT, USHORT, CHAR, UCHAR,
 LLONG, SIZET, IMAX, UMAX, PDIFF, UIPTR,
 DBL, LDBL,
 NOARG,
 MAXSTATE
};

static void pop_arg(union arg *arg, int type, va_list *ap)
{
 switch (type) {
        case PTR: arg->p = __builtin_va_arg(*ap,void *);
 break; case INT: arg->i = __builtin_va_arg(*ap,int);
 break; case UINT: arg->i = __builtin_va_arg(*ap,unsigned int);
 break; case LONG: arg->i = __builtin_va_arg(*ap,long);
 break; case ULONG: arg->i = __builtin_va_arg(*ap,unsigned long);
 break; case ULLONG: arg->i = __builtin_va_arg(*ap,unsigned long long);
 break; case SHORT: arg->i = (short)__builtin_va_arg(*ap,int);
 break; case USHORT: arg->i = (unsigned short)__builtin_va_arg(*ap,int);
 break; case CHAR: arg->i = (signed char)__builtin_va_arg(*ap,int);
 break; case UCHAR: arg->i = (unsigned char)__builtin_va_arg(*ap,int);
 break; case LLONG: arg->i = __builtin_va_arg(*ap,long long);
 break; case SIZET: arg->i = __builtin_va_arg(*ap,size_t);
 break; case IMAX: arg->i = __builtin_va_arg(*ap,intmax_t);
 break; case UMAX: arg->i = __builtin_va_arg(*ap,uintmax_t);
 break; case PDIFF: arg->i = __builtin_va_arg(*ap,ptrdiff_t);
 break; case UIPTR: arg->i = (uintptr_t)__builtin_va_arg(*ap,void *);
 break; case DBL: arg->f = __builtin_va_arg(*ap,double);
 break; case LDBL: arg->f = __builtin_va_arg(*ap,long double);
 }
}