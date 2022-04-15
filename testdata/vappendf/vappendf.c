#include <stdarg.h>
#include <string.h>

typedef unsigned char u8;
typedef unsigned int u32;
typedef unsigned long long u64;
typedef long long i64;

typedef unsigned char etByte;

typedef struct et_info {
  char fmttype;
  etByte base;
  etByte flags;
  etByte type_;
  etByte charset;
  etByte prefix;
} et_info;

static const double arRound[] = {
  5.0e-01, 5.0e-02, 5.0e-03, 5.0e-04, 5.0e-05,
  5.0e-06, 5.0e-07, 5.0e-08, 5.0e-09, 5.0e-10,
};
static const char aDigits[] = "0123456789ABCDEF0123456789abcdef";
static const char aPrefix[] = "-x0\000X0";

static const et_info fmtinfo[] = {
  { 'd', 10, 1, 16, 0, 0 },
  { 's', 0, 4, 5, 0, 0 },
  { 'g', 0, 1, 3, 30, 0 },
  { 'z', 0, 4, 6, 0, 0 },
  { 'q', 0, 4, 9, 0, 0 },
  { 'Q', 0, 4, 10, 0, 0 },
  { 'w', 0, 4, 14, 0, 0 },
  { 'c', 0, 0, 8, 0, 0 },
  { 'o', 8, 0, 0, 0, 2 },
  { 'u', 10, 0, 16, 0, 0 },
  { 'x', 16, 0, 0, 16, 1 },
  { 'X', 16, 0, 0, 0, 4 },

  { 'f', 0, 1, 1, 0, 0 },
  { 'e', 0, 1, 2, 30, 0 },
  { 'E', 0, 1, 2, 14, 0 },
  { 'G', 0, 1, 3, 14, 0 },

  { 'i', 10, 1, 16, 0, 0 },
  { 'n', 0, 0, 4, 0, 0 },
  { '%', 0, 0, 7, 0, 0 },
  { 'p', 16, 0, 13, 0, 1 },

  { 'T', 0, 0, 11, 0, 0 },
  { 'S', 0, 0, 12, 0, 0 },
  { 'r', 10, 1, 15, 0, 0 },
};

typedef struct PrintfArguments {
  int nArg;
  int nUsed;
} PrintfArguments;

typedef struct {
  int flags;
  union {
    char zToken[1];
  } u;
} Expr;

typedef struct Token {
  const char *z;
  unsigned int n;
} Token;

typedef struct Select {
  u8 op;
  u32 selFlags;
  int iLimit, iOffset;
  u32 selId;
  int addrOpenEphm[2];
  Expr *pWhere;
  Expr *pHaving;
  struct Select *pPrior;
  struct Select *pNext;
  Expr *pLimit;
} Select;

typedef struct SrcItem {
  char *zDatabase;
  char *zName;
  char *zAlias;
  Select *pSelect;
  int addrFillSub;
  int regReturn;
  int regResult;
  struct {
    u8 jointype;
    unsigned notIndexed :1;
    unsigned isIndexedBy :1;
    unsigned isTabFunc :1;
    unsigned isCorrelated :1;
    unsigned viaCoroutine :1;
    unsigned isRecursive :1;
    unsigned fromDDL :1;
    unsigned isCte :1;
    unsigned notCte :1;
  } fg;
  int iCursor;
  Expr *pOn;
} SrcItem;

typedef long long sqlite_int64;
typedef unsigned long long sqlite_uint64;

typedef long long sqlite3_int64;
typedef unsigned long long sqlite3_uint64;

typedef struct sqlite3 {
} sqlite3;

typedef struct sqlite3_str {
  sqlite3 *db;
  char *zText;
  u32 nAlloc;
  u32 mxAlloc;
  u32 nChar;
  u8 accError;
  u8 printfFlags;
} sqlite3_str;

static char et_getdigit(long double *val, int *cnt){
  return 0;
}

static sqlite3_int64 getIntArg(PrintfArguments *p){
  return 0;
}

static double getDoubleArg(PrintfArguments *p){
  return 0;
}

static char *getTextArg(PrintfArguments *p){
  return 0;
}

static char *printfTempBuf(sqlite3_str *pAccum, sqlite3_int64 n){
  return 0;
}

static int sqlite3IsNaN(double v) {
  return 0;
}

static int sqlite3DbMallocSize(sqlite3 *db, const void *p){
  return 128;
}

static void sqlite3DbFree(sqlite3* db, void* p) {
}

static void sqlite3RecordErrorByteOffset(sqlite3 *db, const char *z){
}

static void sqlite3RecordErrorOffsetOfExpr(sqlite3 *db, const Expr *pExpr){
}

void sqlite3_str_append(sqlite3_str* pAccum, const char *zIn, int N) {
}

void sqlite3_str_appendf(sqlite3_str* pAccum, const char *zFormat, ...) {
}

void sqlite3_str_appendall(sqlite3_str* pAccum, const char *zIn) {
}

void sqlite3_str_appendchar(sqlite3_str* pAccum, int N, char C) {
}

void sqlite3_str_reset(sqlite3_str* pAccum) {
}

void sqlite3_str_vappendf(
  sqlite3_str *pAccum,
  const char *fmt,
  va_list ap
){
  int c;
  char *bufpt;
  int precision;
  int length;
  int idx;
  int width;
  etByte flag_leftjustify;
  etByte flag_prefix;
  etByte flag_alternateform;
  etByte flag_altform2;
  etByte flag_zeropad;
  etByte flag_long;
  etByte done;
  etByte cThousand;
  etByte xtype = 17;
  u8 bArgList;
  char prefix;
  sqlite_uint64 longvalue;
  long double realvalue;
  const et_info *infop;
  char *zOut;
  int nOut;
  char *zExtra = 0;

  int exp, e2;
  int nsd;
  double rounder;
  etByte flag_dp;
  etByte flag_rtz;

  PrintfArguments *pArgList = 0;
  char buf[70];

  ((void)0);

  bufpt = 0;
  if( (pAccum->printfFlags & 0x02)!=0 ){
    pArgList = __builtin_va_arg(ap, PrintfArguments*);
    bArgList = 1;
  }else{
    bArgList = 0;
  }
  for(; (c=(*fmt))!=0; ++fmt){
    if( c!='%' ){
      bufpt = (char *)fmt;

      do{ fmt++; }while( *fmt && *fmt != '%' );

      sqlite3_str_append(pAccum, bufpt, (int)(fmt - bufpt));
      if( *fmt==0 ) break;
    }
    if( (c=(*++fmt))==0 ){
      sqlite3_str_append(pAccum, "%", 1);
      break;
    }

    flag_leftjustify = flag_prefix = cThousand =
     flag_alternateform = flag_altform2 = flag_zeropad = 0;
    done = 0;
    width = 0;
    flag_long = 0;
    precision = -1;
    do{
      switch( c ){
        case '-': flag_leftjustify = 1; break;
        case '+': flag_prefix = '+'; break;
        case ' ': flag_prefix = ' '; break;
        case '#': flag_alternateform = 1; break;
        case '!': flag_altform2 = 1; break;
        case '0': flag_zeropad = 1; break;
        case ',': cThousand = ','; break;
        default: done = 1; break;
        case 'l': {
          flag_long = 1;
          c = *++fmt;
          if( c=='l' ){
            c = *++fmt;
            flag_long = 2;
          }
          done = 1;
          break;
        }
        case '1': case '2': case '3': case '4': case '5':
        case '6': case '7': case '8': case '9': {
          unsigned wx = c - '0';
          while( (c = *++fmt)>='0' && c<='9' ){
            wx = wx*10 + c - '0';
          }
                                   ;
          width = wx & 0x7fffffff;
          if( c!='.' && c!='l' ){
            done = 1;
          }else{
            fmt--;
          }
          break;
        }
        case '*': {
          if( bArgList ){
            width = (int)getIntArg(pArgList);
          }else{
            width = __builtin_va_arg(ap, int);
          }
          if( width<0 ){
            flag_leftjustify = 1;
            width = width >= -2147483647 ? -width : 0;
          }

          if( (c = fmt[1])!='.' && c!='l' ){
            c = *++fmt;
            done = 1;
          }
          break;
        }
        case '.': {
          c = *++fmt;
          if( c=='*' ){
            if( bArgList ){
              precision = (int)getIntArg(pArgList);
            }else{
              precision = __builtin_va_arg(ap, int);
            }
            if( precision<0 ){
              precision = precision >= -2147483647 ? -precision : -1;
            }
            c = *++fmt;
          }else{
            unsigned px = 0;
            while( c>='0' && c<='9' ){
              px = px*10 + c - '0';
              c = *++fmt;
            }
                                     ;
            precision = px & 0x7fffffff;
          }

          if( c=='l' ){
            --fmt;
          }else{
            done = 1;
          }
          break;
        }
      }
    }while( !done && (c=(*++fmt))!=0 );


    infop = &fmtinfo[0];
    xtype = 17;
    for(idx=0; idx<((int)(sizeof(fmtinfo)/sizeof(fmtinfo[0]))); idx++){
      if( c==fmtinfo[idx].fmttype ){
        infop = &fmtinfo[idx];
        xtype = infop->type_;
        break;
      }
    }
    ((void)0);
    ((void)0);
    switch( xtype ){
      case 13:
        flag_long = sizeof(char*)==sizeof(i64) ? 2 :
                     sizeof(char*)==sizeof(long int) ? 1 : 0;

      case 15:
      case 0:
        cThousand = 0;

      case 16:
        if( infop->flags & 1 ){
          i64 v;
          if( bArgList ){
            v = getIntArg(pArgList);
          }else if( flag_long ){
            if( flag_long==2 ){
              v = __builtin_va_arg(ap, i64) ;
            }else{
              v = __builtin_va_arg(ap, long int);
            }
          }else{
            v = __builtin_va_arg(ap, int);
          }
          if( v<0 ){
                                         ;
                               ;
            longvalue = ~v;
            longvalue++;
            prefix = '-';
          }else{
            longvalue = v;
            prefix = flag_prefix;
          }
        }else{
          if( bArgList ){
            longvalue = (u64)getIntArg(pArgList);
          }else if( flag_long ){
            if( flag_long==2 ){
              longvalue = __builtin_va_arg(ap, u64);
            }else{
              longvalue = __builtin_va_arg(ap, unsigned long int);
            }
          }else{
            longvalue = __builtin_va_arg(ap, unsigned int);
          }
          prefix = 0;
        }
        if( longvalue==0 ) flag_alternateform = 0;
        if( flag_zeropad && precision<width-(prefix!=0) ){
          precision = width-(prefix!=0);
        }
        if( precision<70 -10-70/3 ){
          nOut = 70;
          zOut = buf;
        }else{
          u64 n;
          n = (u64)precision + 10;
          if( cThousand ) n += precision/3;
          zOut = zExtra = printfTempBuf(pAccum, n);
          if( zOut==0 ) return;
          nOut = (int)n;
        }
        bufpt = &zOut[nOut-1];
        if( xtype==15 ){
          static const char zOrd[] = "thstndrd";
          int x = (int)(longvalue % 10);
          if( x>=4 || (longvalue/10)%10==1 ){
            x = 0;
          }
          *(--bufpt) = zOrd[x*2+1];
          *(--bufpt) = zOrd[x*2];
        }
        {
          const char *cset = &aDigits[infop->charset];
          u8 base = infop->base;
          do{
            *(--bufpt) = cset[longvalue%base];
            longvalue = longvalue/base;
          }while( longvalue>0 );
        }
        length = (int)(&zOut[nOut-1]-bufpt);
        while( precision>length ){
          *(--bufpt) = '0';
          length++;
        }
        if( cThousand ){
          int nn = (length - 1)/3;
          int ix = (length - 1)%3 + 1;
          bufpt -= nn;
          for(idx=0; nn>0; idx++){
            bufpt[idx] = bufpt[idx+nn];
            ix--;
            if( ix==0 ){
              bufpt[++idx] = cThousand;
              nn--;
              ix = 3;
            }
          }
        }
        if( prefix ) *(--bufpt) = prefix;
        if( flag_alternateform && infop->prefix ){
          const char *pre;
          char x;
          pre = &aPrefix[infop->prefix];
          for(; (x=(*pre))!=0; pre++) *(--bufpt) = x;
        }
        length = (int)(&zOut[nOut-1]-bufpt);
        break;
      case 1:
      case 2:
      case 3:
        if( bArgList ){
          realvalue = getDoubleArg(pArgList);
        }else{
          realvalue = __builtin_va_arg(ap, double);
        }

        if( precision<0 ) precision = 6;

        if( precision>100000000 ){
          precision = 100000000;
        }

        if( realvalue<0.0 ){
          realvalue = -realvalue;
          prefix = '-';
        }else{
          prefix = flag_prefix;
        }
        if( xtype==3 && precision>0 ) precision--;
                                   ;
        idx = precision & 0xfff;
        rounder = arRound[idx%10];
        while( idx>=10 ){ rounder *= 1.0e-10; idx -= 10; }
        if( xtype==1 ){
          double rx = (double)realvalue;
          sqlite3_uint64 u;
          int ex;
          __builtin___memcpy_chk (&u, &rx, sizeof(u), __builtin_object_size (&u, 0));
          ex = -1023 + (int)((u>>52)&0x7ff);
          if( precision+(ex/3) < 15 ) rounder += realvalue*3e-16;
          realvalue += rounder;
        }

        exp = 0;
        if( sqlite3IsNaN((double)realvalue) ){
          bufpt = "NaN";
          length = 3;
          break;
        }
        if( realvalue>0.0 ){
          long double scale = 1.0;
          while( realvalue>=1e100*scale && exp<=350 ){ scale *= 1e100;exp+=100;}
          while( realvalue>=1e10*scale && exp<=350 ){ scale *= 1e10; exp+=10; }
          while( realvalue>=10.0*scale && exp<=350 ){ scale *= 10.0; exp++; }
          realvalue /= scale;
          while( realvalue<1e-8 ){ realvalue *= 1e8; exp-=8; }
          while( realvalue<1.0 ){ realvalue *= 10.0; exp--; }
          if( exp>350 ){
            bufpt = buf;
            buf[0] = prefix;
            __builtin___memcpy_chk (buf+(prefix!=0), "Inf",4, __builtin_object_size (buf+(prefix!=0), 0));
            length = 3+(prefix!=0);
            break;
          }
        }
        bufpt = buf;

        if( xtype!=1 ){
          realvalue += rounder;
          if( realvalue>=10.0 ){ realvalue *= 0.1; exp++; }
        }
        if( xtype==3 ){
          flag_rtz = !flag_alternateform;
          if( exp<-4 || exp>precision ){
            xtype = 2;
          }else{
            precision = precision - exp;
            xtype = 1;
          }
        }else{
          flag_rtz = flag_altform2;
        }
        if( xtype==2 ){
          e2 = 0;
        }else{
          e2 = exp;
        }
        {
          i64 szBufNeeded;
          szBufNeeded = ((e2)>(0)?(e2):(0))+(i64)precision+(i64)width+15;
          if( szBufNeeded > 70 ){
            bufpt = zExtra = printfTempBuf(pAccum, szBufNeeded);
            if( bufpt==0 ) return;
          }
        }
        zOut = bufpt;
        nsd = 16 + flag_altform2*10;
        flag_dp = (precision>0 ?1:0) | flag_alternateform | flag_altform2;

        if( prefix ){
          *(bufpt++) = prefix;
        }

        if( e2<0 ){
          *(bufpt++) = '0';
        }else{
          for(; e2>=0; e2--){
            *(bufpt++) = et_getdigit(&realvalue,&nsd);
          }
        }

        if( flag_dp ){
          *(bufpt++) = '.';
        }


        for(e2++; e2<0; precision--, e2++){
          ((void)0);
          *(bufpt++) = '0';
        }

        while( (precision--)>0 ){
          *(bufpt++) = et_getdigit(&realvalue,&nsd);
        }

        if( flag_rtz && flag_dp ){
          while( bufpt[-1]=='0' ) *(--bufpt) = 0;
          ((void)0);
          if( bufpt[-1]=='.' ){
            if( flag_altform2 ){
              *(bufpt++) = '0';
            }else{
              *(--bufpt) = 0;
            }
          }
        }

        if( xtype==2 ){
          *(bufpt++) = aDigits[infop->charset];
          if( exp<0 ){
            *(bufpt++) = '-'; exp = -exp;
          }else{
            *(bufpt++) = '+';
          }
          if( exp>=100 ){
            *(bufpt++) = (char)((exp/100)+'0');
            exp %= 100;
          }
          *(bufpt++) = (char)(exp/10+'0');
          *(bufpt++) = (char)(exp%10+'0');
        }
        *bufpt = 0;

        length = (int)(bufpt-zOut);
        bufpt = zOut;

        if( flag_zeropad && !flag_leftjustify && length < width){
          int i;
          int nPad = width - length;
          for(i=width; i>=nPad; i--){
            bufpt[i] = bufpt[i-nPad];
          }
          i = prefix!=0;
          while( nPad-- ) bufpt[i++] = '0';
          length = width;
        }

        break;
      case 4:
        if( !bArgList ){
          *(__builtin_va_arg(ap, int*)) = pAccum->nChar;
        }
        length = width = 0;
        break;
      case 7:
        buf[0] = '%';
        bufpt = buf;
        length = 1;
        break;
      case 8:
        if( bArgList ){
          bufpt = getTextArg(pArgList);
          length = 1;
          if( bufpt ){
            buf[0] = c = *(bufpt++);
            if( (c&0xc0)==0xc0 ){
              while( length<4 && (bufpt[0]&0xc0)==0x80 ){
                buf[length++] = *(bufpt++);
              }
            }
          }else{
            buf[0] = 0;
          }
        }else{
          unsigned int ch = __builtin_va_arg(ap, unsigned int);
          if( ch<0x00080 ){
            buf[0] = ch & 0xff;
            length = 1;
          }else if( ch<0x00800 ){
            buf[0] = 0xc0 + (u8)((ch>>6)&0x1f);
            buf[1] = 0x80 + (u8)(ch & 0x3f);
            length = 2;
          }else if( ch<0x10000 ){
            buf[0] = 0xe0 + (u8)((ch>>12)&0x0f);
            buf[1] = 0x80 + (u8)((ch>>6) & 0x3f);
            buf[2] = 0x80 + (u8)(ch & 0x3f);
            length = 3;
          }else{
            buf[0] = 0xf0 + (u8)((ch>>18) & 0x07);
            buf[1] = 0x80 + (u8)((ch>>12) & 0x3f);
            buf[2] = 0x80 + (u8)((ch>>6) & 0x3f);
            buf[3] = 0x80 + (u8)(ch & 0x3f);
            length = 4;
          }
        }
        if( precision>1 ){
          width -= precision-1;
          if( width>1 && !flag_leftjustify ){
            sqlite3_str_appendchar(pAccum, width-1, ' ');
            width = 0;
          }
          while( precision-- > 1 ){
            sqlite3_str_append(pAccum, buf, length);
          }
        }
        bufpt = buf;
        flag_altform2 = 1;
        goto adjust_width_for_utf8;
      case 5:
      case 6:
        if( bArgList ){
          bufpt = getTextArg(pArgList);
          xtype = 5;
        }else{
          bufpt = __builtin_va_arg(ap, char*);
        }
        if( bufpt==0 ){
          bufpt = "";
        }else if( xtype==6 ){
          if( pAccum->nChar==0
           && pAccum->mxAlloc
           && width==0
           && precision<0
           && pAccum->accError==0
          ){
            ((void)0);
            pAccum->zText = bufpt;
            pAccum->nAlloc = sqlite3DbMallocSize(pAccum->db, bufpt);
            pAccum->nChar = 0x7fffffff & (int)strlen(bufpt);
            pAccum->printfFlags |= 0x04;
            length = 0;
            break;
          }
          zExtra = bufpt;
        }
        if( precision>=0 ){
          if( flag_altform2 ){
            unsigned char *z = (unsigned char*)bufpt;
            while( precision-- > 0 && z[0] ){
              { if( (*(z++))>=0xc0 ){ while( (*z & 0xc0)==0x80 ){ z++; } } };
            }
            length = (int)(z - (unsigned char*)bufpt);
          }else{
            for(length=0; length<precision && bufpt[length]; length++){}
          }
        }else{
          length = 0x7fffffff & (int)strlen(bufpt);
        }
      adjust_width_for_utf8:
        if( flag_altform2 && width>0 ){
          int ii = length - 1;
          while( ii>=0 ) if( (bufpt[ii--] & 0xc0)==0x80 ) width++;
        }
        break;
      case 9:
      case 10:
      case 14: {
        int i, j, k, n, isnull;
        int needQuote;
        char ch;
        char q = ((xtype==14)?'"':'\'');
        char *escarg;

        if( bArgList ){
          escarg = getTextArg(pArgList);
        }else{
          escarg = __builtin_va_arg(ap, char*);
        }
        isnull = escarg==0;
        if( isnull ) escarg = (xtype==10 ? "NULL" : "(NULL)");
        k = precision;
        for(i=n=0; k!=0 && (ch=escarg[i])!=0; i++, k--){
          if( ch==q ) n++;
          if( flag_altform2 && (ch&0xc0)==0xc0 ){
            while( (escarg[i+1]&0xc0)==0x80 ){ i++; }
          }
        }
        needQuote = !isnull && xtype==10;
        n += i + 3;
        if( n>70 ){
          bufpt = zExtra = printfTempBuf(pAccum, n);
          if( bufpt==0 ) return;
        }else{
          bufpt = buf;
        }
        j = 0;
        if( needQuote ) bufpt[j++] = q;
        k = i;
        for(i=0; i<k; i++){
          bufpt[j++] = ch = escarg[i];
          if( ch==q ) bufpt[j++] = ch;
        }
        if( needQuote ) bufpt[j++] = q;
        bufpt[j] = 0;
        length = j;
        goto adjust_width_for_utf8;
      }
      case 11: {
        if( (pAccum->printfFlags & 0x01)==0 ) return;
        if( flag_alternateform ){
          Expr *pExpr = __builtin_va_arg(ap, Expr*);
          if( (pExpr) && (!(((pExpr)->flags&(0x000400))!=0)) ){
            sqlite3_str_appendall(pAccum, (const char*)pExpr->u.zToken);
            sqlite3RecordErrorOffsetOfExpr(pAccum->db, pExpr);
          }
        }else{
          Token *pToken = __builtin_va_arg(ap, Token*);
          ((void)0);
          if( pToken && pToken->n ){
            sqlite3_str_append(pAccum, (const char*)pToken->z, pToken->n);
            sqlite3RecordErrorByteOffset(pAccum->db, pToken->z);
          }
        }
        length = width = 0;
        break;
      }
      case 12: {
        int i = 0;
        SrcItem *pItem;
        (void)i;
        if( (pAccum->printfFlags & 0x01)==0 ) return;
        pItem = __builtin_va_arg(ap, SrcItem*);
        ((void)0);
        if( pItem->zAlias && !flag_altform2 ){
          sqlite3_str_appendall(pAccum, pItem->zAlias);
        }else if( pItem->zName ){
          if( pItem->zDatabase ){
            sqlite3_str_appendall(pAccum, pItem->zDatabase);
            sqlite3_str_append(pAccum, ".", 1);
          }
          sqlite3_str_appendall(pAccum, pItem->zName);
        }else if( pItem->zAlias ){
          sqlite3_str_appendall(pAccum, pItem->zAlias);
        }else if( (pItem->pSelect) ){
          sqlite3_str_appendf(pAccum, "SUBQUERY %u", pItem->pSelect->selId);
        }
        length = width = 0;
        break;
      }
      default: {
        ((void)0);
        return;
      }
    }
    width -= length;
    if( width>0 ){
      if( !flag_leftjustify ) sqlite3_str_appendchar(pAccum, width, ' ');
      sqlite3_str_append(pAccum, bufpt, length);
      if( flag_leftjustify ) sqlite3_str_appendchar(pAccum, width, ' ');
    }else{
      sqlite3_str_append(pAccum, bufpt, length);
    }

    if( zExtra ){
      sqlite3DbFree(pAccum->db, zExtra);
      zExtra = 0;
    }
  }
}
