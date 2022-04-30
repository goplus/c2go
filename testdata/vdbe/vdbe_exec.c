typedef long i64;
typedef unsigned long u64;
typedef unsigned int u32;
typedef unsigned short int u16;
typedef short int i16;
typedef unsigned char u8;
typedef signed char i8;

typedef struct sqlite3 sqlite3;
struct sqlite3 {
  struct Vdbe *pVdbe;
  int nDb;
  u32 mDbFlags;
  u64 flags;
  i64 lastRowid;
  i64 szMmap;
  u32 nSchemaLock;
  unsigned int openFlags;
  int errCode;
  int errByteOffset;
  int errMask;
  int iSysErrno;
  u32 dbOptFlags;
  u8 enc;
  u8 autoCommit;
  u8 temp_store;
  u8 mallocFailed;
  u8 bBenignMalloc;
  u8 dfltLockMode;
  signed char nextAutovac;
  u8 suppressErr;
  u8 vtabOnConflict;
  u8 isTransactionSavepoint;
  u8 mTrace;
  u8 noSharedCache;
  u8 nSqlExec;
  u8 eOpenState;
  int nextPagesize;
  i64 nChange;
  i64 nTotalChange;
  int aLimit[(11 +1)];
  int nMaxSorterMmap;
  struct sqlite3InitInfo {
    u8 iDb;
    u8 busy;
    unsigned orphanTrigger : 1;
    unsigned imposterTable : 1;
    unsigned reopenMemdb : 1;
    const char **azInit;
  } init;
  int nVdbeActive;
  int nVdbeRead;
  int nVdbeWrite;
  int nVdbeExec;
  int nVDestroy;
  int nExtension;
  void **aExtension;
  union {
    void (*xLegacy)(void*,const char*);
    int (*xV2)(u32,void*,void*,void*);
  } trace;
  void *pTraceArg;

  void (*xProfile)(void*,const char*,u64);
  void *pProfileArg;

  void *pCommitArg;
  int (*xCommitCallback)(void*);
  void *pRollbackArg;
  void (*xRollbackCallback)(void*);
  void *pUpdateArg;
  void *pAutovacPagesArg;
  void (*xAutovacDestr)(void*);
  unsigned int (*xAutovacPages)(void*,const char*,u32,u32,u32);
  int (*xWalCallback)(void *, sqlite3 *, const char *, int);
  void *pWalArg;

  void(*xCollNeeded)(void*,sqlite3*,int eTextRep,const char*);
  void(*xCollNeeded16)(void*,sqlite3*,int eTextRep,const void*);
  void *pCollNeededArg;
  union {
    volatile int isInterrupted;
    double notUsed1;
  } u1;

  void *pAuthArg;

  int (*xProgress)(void *);
  void *pProgressArg;
  unsigned nProgressOps;

  int nVTrans;
  int nAnalysisLimit;
  int busyTimeout;
  int nSavepoint;
  int nStatement;
  i64 nDeferredCons;
  i64 nDeferredImmCons;
  int *pnBytesFreed;
};

typedef struct sqlite3_value {
  union MemValue {
    double r;
    i64 i;
    int nZero;
    const char *zPType;
  } u;
  u16 flags;
  u8 enc;
  u8 eSubtype;
  int n;
  char *z;

  char *zMalloc;
  int szMalloc;
  u32 uTemp;
  sqlite3 *db;
  void (*xDel)(void*);
} Mem;

typedef struct VdbeOp {
  u8 opcode;
  signed char p4type;
  u16 p5;
  int p1;
  int p2;
  int p3;
  union p4union {
    int i;
    void *p;
    char *z;
    i64 *pI64;
    double *pReal;
    Mem *pMem;
    u32 *ai;
  } p4;
} VdbeOp, Op;

typedef struct VdbeFrame VdbeFrame;
typedef struct Vdbe Vdbe;

struct Vdbe {
  sqlite3 *db;
  Vdbe *pPrev,*pNext;
  u32 iVdbeMagic;
  int nMem;
  int nCursor;
  u32 cacheCtr;
  int pc;
  int rc;
  i64 nChange;
  int iStatement;
  i64 iCurrentTime;
  i64 nFkConstraint;
  i64 nStmtDefCons;
  i64 nStmtDefImmCons;
  Mem *aMem;
  Mem **apArg;
  Mem *aVar;

  Op *aOp;
  int nOp;
  int nOpAlloc;
  Mem *aColName;
  Mem *pResultSet;
  char *zErrMsg;

  i64 startTime;

  u16 nResColumn;
  u8 errorAction;
  u8 minWriteFileFormat;
  u8 prepFlags;
  u8 doingRerun;
  u32 aCounter[9];
  char *zSql;

  void *pFree;
  VdbeFrame *pFrame;
  VdbeFrame *pDelFrame;
  int nFrame;
  u32 expmask;
};

struct VdbeFrame {
  Vdbe *v;
  VdbeFrame *pParent;
  Op *aOp;
  i64 *anExec;
  Mem *aMem;
  u8 *aOnce;
  void *token;
  i64 lastRowid;

  int nCursor;
  int pc;
  int nOp;
  int nMem;
  int nChildMem;
  int nChildCsr;
  i64 nChange;
  i64 nDbChange;
};

static int sqlite3VdbeExec(
  Vdbe *p
){
  Op *aOp = p->aOp;
  Op *pOp = aOp;
  int rc = 0;
  sqlite3 *db = p->db;
  u8 encoding = ((db)->enc);
  int iCompare = 0;
  u64 nVmStep = 0;
  u64 nProgressLimit;

  Mem *aMem = p->aMem;
  Mem *pIn1 = 0;
  Mem *pIn2 = 0;
  Mem *pIn3 = 0;
  Mem *pOut = 0;

  if( db->xProgress ){
    u32 iPrior = p->aCounter[4];
    nProgressLimit = db->nProgressOps - (iPrior % db->nProgressOps);
  }else{
    nProgressLimit = (0xffffffff|(((u64)0xffffffff)<<32));
  }

  if( p->rc==7 ){
    goto no_mem;
  }
  p->rc = 0;
  p->iCurrentTime = 0;
  p->pResultSet = 0;
  if( 0 ) goto abort_due_to_interrupt;

  for(pOp=&aOp[p->pc]; 1; pOp++){
    nVmStep++;
    switch( pOp->opcode ){
case 11: {
  pOp = &aOp[pOp->p2 - 1];

  if( 0 ) goto abort_due_to_interrupt;

  while( nVmStep>=nProgressLimit && db->xProgress!=0 ){
    ((void)0);
    nProgressLimit += db->nProgressOps;
    if( db->xProgress(db->pProgressArg) ){
      nProgressLimit = (0xffffffff|(((u64)0xffffffff)<<32));
      rc = 9;
      goto abort_due_to_error;
    }
  }
  break;
}

case 12: {
  pIn1 = &aMem[pOp->p1];
  pIn1->flags = 0x0004;
  pIn1->u.i = (int)(pOp-aOp);
jump_to_p2:
  pOp = &aOp[pOp->p2 - 1];
  break;
}

case 67: {
  pIn1 = &aMem[pOp->p1];
  ((void)0);
  pOp = &aOp[pIn1->u.i];
  pIn1->flags = 0x0080;
  break;
}

case 13: {
  pOut = &aMem[pOp->p1];
  pOut->u.i = pOp->p3 - 1;
  pOut->flags = 0x0004;
  if( pOp->p2 ) goto jump_to_p2;
  break;
}

case 68: {
  VdbeOp *pCaller;
  pIn1 = &aMem[pOp->p1];
  pCaller = &aOp[pIn1->u.i];
  pOp = &aOp[pCaller->p2 - 1];
  pIn1->flags = 0x0080;
  break;
}

case 14: {
  int pcDest;
  pIn1 = &aMem[pOp->p1];
  pIn1->flags = 0x0004;
  pcDest = (int)pIn1->u.i;
  pIn1->u.i = (int)(pOp - aOp);
  pOp = &aOp[pcDest];
  break;
}

case 69: {
  pIn3 = &aMem[pOp->p3];
  if( (pIn3->flags & 0x0001)==0 ) break;
}

case 70: {
  VdbeFrame *pFrame;
  int pcx;
  pcx = (int)(pOp - aOp);
  if( pOp->p1==0 && p->pFrame ){
    pFrame = p->pFrame;
    p->pFrame = pFrame->pParent;
    p->nFrame--;
    if( pOp->p2==4 ){
      pcx = p->aOp[pcx].p2-1;
    }
    aOp = p->aOp;
    aMem = p->aMem;
    pOp = &aOp[pcx];
    break;
  }
  p->rc = pOp->p1;
  p->errorAction = (u8)pOp->p2;
  p->pc = pcx;
  if( rc==5 ){
    p->rc = 5;
  }else{
    rc = p->rc ? 1 : 101;
  }
  goto vdbe_return;
}

case 71: {
  pOut->u.i = pOp->p1;
  break;
}

case 72: {
  pOut->u.i = *pOp->p4.pI64;
  break;
}

case 153: {
  pOut->flags = 0x0008;
  pOut->u.r = *pOp->p4.pReal;
  break;
}

case 117: {
  if( encoding!=1 ){
    if( rc ) goto too_big;
    if( 0 ) goto no_mem;
    pOut->szMalloc = 0;
    pOut->flags |= 0x0800;
    pOp->p4type = (-7);
    pOp->p4.z = pOut->z;
    pOp->p1 = pOut->n;
  }

  if( pOp->p1>db->aLimit[0] ){
    goto too_big;
  }
  pOp->opcode = 73;
}

case 73: {
  pOut->flags = 0x0002|0x0800|0x0200;
  pOut->z = pOp->p4.z;
  pOut->n = pOp->p1;
  pOut->enc = encoding;
  if( pOp->p3>0 ){
    pIn3 = &aMem[pOp->p3];
    if( pIn3->u.i==pOp->p5 ) pOut->flags = 0x0010|0x0800|0x0200;
  }
  break;
}

case 74: {
  int cnt;
  u16 nullFlag;
  cnt = pOp->p3-pOp->p2;
  pOut->flags = nullFlag = pOp->p1 ? (0x0001|0x0100) : 0x0001;
  pOut->n = 0;
  while( cnt>0 ){
    pOut++;
    pOut->flags = nullFlag;
    pOut->n = 0;
    cnt--;
  }
  break;
}

case 75: {
  pOut = &aMem[pOp->p1];
  pOut->flags = (pOut->flags&~(0x0080|0x003f))|0x0001;
  break;
}

case 76: {
  if( pOp->p4.z==0 ){
    if( 0 ) goto no_mem;
  }else{
  }
  pOut->enc = encoding;
  break;
}

case 77: {
  Mem *pVar;
  (void)pVar;
  if( 0 ){
    goto too_big;
  }
  pOut = &aMem[pOp->p2];
  pOut->flags &= ~(0x0400|0x1000);
  pOut->flags |= 0x0800|0x0040;
  break;
}

case 78: {
  int n;
  int p1;
  int p2;

  n = pOp->p3;
  p1 = pOp->p1;
  p2 = pOp->p2;
  ((void)0);
  ((void)0);

  pIn1 = &aMem[p1];
  pOut = &aMem[p2];
  do{
    if( 0 ){ goto no_mem;};
    pIn1++;
    pOut++;
  }while( --n );
  break;
}

case 79: {
  int n;

  n = pOp->p3;
  pIn1 = &aMem[pOp->p1];
  pOut = &aMem[pOp->p2];
  while( 1 ){
    if( 0 ){ goto no_mem;};
    if( (n--)==0 ) break;
    pOut++;
    pIn1++;
  }
  break;
}

case 80: {
  pIn1 = &aMem[pOp->p1];
  pOut = &aMem[pOp->p2];
  break;
}

case 81: {
  pIn1 = &aMem[pOp->p1];
  pOut = &aMem[pOp->p2];
  break;
}

case 82: {
  if( 0 ){
    goto abort_due_to_error;
  }
  break;
}

case 83: {
  Mem *pMem;
  int i;
  p->cacheCtr = (p->cacheCtr + 2)|1;
  pMem = p->pResultSet = &aMem[pOp->p1];
  (void)pMem;
  for(i=0; i<pOp->p2; i++){
    if( 0 ){ goto no_mem;};
  }
  if( db->mallocFailed ) goto no_mem;
  if( db->mTrace & 0x04 ){
    db->trace.xV2(0x04, db->pTraceArg, p, 0);
  }
  p->pc = (int)(pOp - aOp) + 1;
  rc = 100;
  goto vdbe_return;
}

case 111: {
  i64 nByte;
  u16 flags1;
  u16 flags2;

  pIn1 = &aMem[pOp->p1];
  pIn2 = &aMem[pOp->p2];
  pOut = &aMem[pOp->p3];

  flags1 = pIn1->flags;
  if( (flags1 | pIn2->flags) & 0x0001 ){
    break;
  }
  if( (flags1 & (0x0002|0x0010))==0 ){
    if( 0 ) goto no_mem;
    flags1 = pIn1->flags & ~0x0002;
  }else if( (flags1 & 0x4000)!=0 ){
    if( 0 ) goto no_mem;
    flags1 = pIn1->flags & ~0x0002;
  }
  flags2 = pIn2->flags;
  if( (flags2 & (0x0002|0x0010))==0 ){
    if( 0 ) goto no_mem;
    flags2 = pIn2->flags & ~0x0002;
  }else if( (flags2 & 0x4000)!=0 ){
    if( 0 ) goto no_mem;
    flags2 = pIn2->flags & ~0x0002;
  }
  nByte = pIn1->n + pIn2->n;
  if( nByte>db->aLimit[0] ){
    goto too_big;
  }
  if( 0 ){
    goto no_mem;
  }
  ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0002);
  if( pOut!=pIn2 ){
    pIn2->flags = flags2;
  }
  pIn1->flags = flags1;
  pOut->z[nByte]=0;
  pOut->z[nByte+1] = 0;
  pOut->z[nByte+2] = 0;
  pOut->flags |= 0x0200;
  pOut->n = (int)nByte;
  pOut->enc = encoding;
  break;
}

case 106:
case 107:
case 108:
case 109:
case 110: {
  u16 flags;
  u16 type1;
  u16 type2;
  i64 iA;
  i64 iB;
  double rA;
  double rB;

  pIn1 = &aMem[pOp->p1];
  pIn2 = &aMem[pOp->p2];
  pOut = &aMem[pOp->p3];
  flags = pIn1->flags | pIn2->flags;
  if( (type1 & type2 & 0x0004)!=0 ){
    iA = pIn1->u.i;
    iB = pIn2->u.i;
    switch( pOp->opcode ){
      case 106: if( 1 ) goto fp_math; break;
      case 107: if( 1 ) goto fp_math; break;
      case 108: if( 1 ) goto fp_math; break;
      case 109: {
        if( iA==0 ) goto arithmetic_result_is_null;
        if( iA==-1 && iB==(((i64)-1) - (0xffffffff|(((i64)0x7fffffff)<<32))) ) goto fp_math;
        iB /= iA;
        break;
      }
      default: {
        if( iA==0 ) goto arithmetic_result_is_null;
        if( iA==-1 ) iA = 1;
        iB %= iA;
        break;
      }
    }
    pOut->u.i = iB;
    ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0004);
  }else if( (flags & 0x0001)!=0 ){
    goto arithmetic_result_is_null;
  }else{
fp_math:
    switch( pOp->opcode ){
      case 106: rB += rA; break;
      case 107: rB -= rA; break;
      case 108: rB *= rA; break;
      case 109: {
        if( rA==(double)0 ) goto arithmetic_result_is_null;
        rB /= rA;
        break;
      }
      default: {
        if( iA==0 ) goto arithmetic_result_is_null;
        if( iA==-1 ) iA = 1;
        rB = (double)(iB % iA);
        break;
      }
    }
    if( 0 ){
      goto arithmetic_result_is_null;
    }
    pOut->u.r = rB;
    ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0008);
  }
  break;
arithmetic_result_is_null:
  break;
}

case 84: {
  break;
}

case 102:
case 103:
case 104:
case 105: {
  i64 iA;
  u64 uA;
  i64 iB;
  u8 op;
  pIn1 = &aMem[pOp->p1];
  pIn2 = &aMem[pOp->p2];
  pOut = &aMem[pOp->p3];
  if( (pIn1->flags | pIn2->flags) & 0x0001 ){
    break;
  }
  op = pOp->opcode;
  if( op==102 ){
    iA &= iB;
  }else if( op==103 ){
    iA |= iB;
  }else if( iB!=0 ){
    if( iB<0 ){
      ((void)0);
      op = 2*104 + 1 - op;
      iB = iB>(-64) ? -iB : 64;
    }
    if( iB>=64 ){
      iA = (iA>=0 || op==104) ? 0 : -1;
    }else{
      if( op==104 ){
        uA <<= iB;
      }else{
        uA >>= iB;
        if( iA<0 ) uA |= ((((u64)0xffffffff)<<32)|0xffffffff) << (64-iB);
      }
    }
  }
  pOut->u.i = iA;
  ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0004);
  break;
}

case 85: {
  pIn1 = &aMem[pOp->p1];
  pIn1->u.i += pOp->p2;
  break;
}

case 15: {
  pIn1 = &aMem[pOp->p1];
  if( (pIn1->flags & 0x0004)==0 ){
    if( (pIn1->flags & 0x0004)==0 ){
      if( pOp->p2==0 ){
        rc = 20;
        goto abort_due_to_error;
      }else{
        goto jump_to_p2;
      }
    }
  }
  ((pIn1)->flags = ((pIn1)->flags&~(0xc1bf|0x4000))|0x0004);
  break;
}

case 86: {
  pIn1 = &aMem[pOp->p1];
  break;
}

case 87: {
  pIn1 = &aMem[pOp->p1];
  if( rc ) goto abort_due_to_error;
  break;
}

case 53:
case 52:
case 56:
case 55:
case 54:
case 57: {
  int res, res2;
  char affinity;
  u16 flags1;
  u16 flags3;

  pIn1 = &aMem[pOp->p1];
  pIn3 = &aMem[pOp->p3];
  flags1 = pIn1->flags;
  flags3 = pIn3->flags;
  if( (flags1 & flags3 & 0x0004)!=0 ){
    if( pIn3->u.i > pIn1->u.i ){
      iCompare = +1;
      if( 1 ){
        goto jump_to_p2;
      }
    }else if( pIn3->u.i < pIn1->u.i ){
      iCompare = -1;
      if( 1 ){
        goto jump_to_p2;
      }
    }else{
      iCompare = 0;
      if( 1 ){
        goto jump_to_p2;
      }
    }
    break;
  }
  if( (flags1 | flags3)&0x0001 ){
    if( pOp->p5 & 0x80 ){
      if( (flags1&flags3&0x0001)!=0
       && (flags3&0x0100)==0
      ){
        res = 0;
      }else{
        res = ((flags3 & 0x0001) ? -1 : +1);
      }
    }else{
      iCompare = 1;
      if( pOp->p5 & 0x10 ){
        goto jump_to_p2;
      }
      break;
    }
  }else{
    affinity = pOp->p5 & 0x47;
    if( affinity>=0x43 ){
      if( (flags1 | flags3)&0x0002 ){
        if( (flags1 & (0x0004|0x0020|0x0008|0x0002))==0x0002 ){
          flags3 = pIn3->flags;
        }
      }
    }else if( affinity==0x42 ){
      if( (flags1 & 0x0002)==0 && (flags1&(0x0004|0x0008|0x0020))!=0 ){
        flags1 = (pIn1->flags & ~0xc1bf) | (flags1 & 0xc1bf);
        if( pIn1==pIn3 ) flags3 = flags1 | 0x0002;
      }
      if( (flags3 & 0x0002)==0 && (flags3&(0x0004|0x0008|0x0020))!=0 ){
        flags3 = (pIn3->flags & ~0xc1bf) | (flags3 & 0xc1bf);
      }
    }
  }
  iCompare = res;
  pIn3->flags = flags3;
  pIn1->flags = flags1;
  if( res2 ){
    goto jump_to_p2;
  }
  break;
}

case 58: {
  if( iCompare==0 ) goto jump_to_p2;
  break;
}

case 88: {
  break;
}

case 89: {
  int n;
  int i;
  int p1;
  int p2;
  u32 idx;
  int bRev;
  u32 *aPermute;
  (void)p2;

  if( (pOp->p5 & 0x01)==0 ){
    aPermute = 0;
  }else{
    aPermute = pOp[-1].p4.ai + 1;
    ((void)0);
  }
  n = pOp->p3;
  p1 = pOp->p1;
  p2 = pOp->p2;
  for(i=0; i<n; i++){
    idx = aPermute ? aPermute[i] : (u32)i;
    if( iCompare ){
      if(((aMem[p1+idx].flags & 0x0001) || (aMem[p2+idx].flags & 0x0001))
      ){
        iCompare = -iCompare;
      }
      if( bRev ) iCompare = -iCompare;
      break;
    }
  }
  break;
}

case 16: {
  if( iCompare<0 ){
                        ; pOp = &aOp[pOp->p1 - 1];
  }else if( iCompare==0 ){
                        ; pOp = &aOp[pOp->p2 - 1];
  }else{
                        ; pOp = &aOp[pOp->p3 - 1];
  }
  break;
}

case 44:
case 43: {
  int v1;
  int v2;
  if( pOp->opcode==44 ){
    static const unsigned char and_logic[] = { 0, 0, 0, 0, 1, 2, 0, 2, 2 };
    v1 = and_logic[v1*3+v2];
  }else{
    static const unsigned char or_logic[] = { 0, 1, 2, 1, 1, 1, 2, 1, 2 };
    v1 = or_logic[v1*3+v2];
  }
  pOut = &aMem[pOp->p3];
  if( v1==2 ){
    ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0001);
  }else{
    pOut->u.i = v1;
    ((pOut)->flags = ((pOut)->flags&~(0xc1bf|0x4000))|0x0004);
  }
  break;
}

case 90: {
  break;
}

case 19: {
  pIn1 = &aMem[pOp->p1];
  pOut = &aMem[pOp->p2];
  break;
}

case 114: {
  pIn1 = &aMem[pOp->p1];
  pOut = &aMem[pOp->p2];
  if( (pIn1->flags & 0x0001)==0 ){
    pOut->flags = 0x0004;
  }
  break;
}

case 17: {
  u32 iAddr;
  ((void)0);
  if( p->pFrame ){
    iAddr = (int)(pOp - p->aOp);
    if( (p->pFrame->aOnce[iAddr/8] & (1<<(iAddr & 7)))!=0 ){
      goto jump_to_p2;
    }
    p->pFrame->aOnce[iAddr/8] |= 1<<(iAddr & 7);
  }else{
    if( p->aOp[0].p1==pOp->p1 ){
      goto jump_to_p2;
    }
  }
  pOp->p1 = p->aOp[0].p1;
  break;
}

case 50: {
  pIn1 = &aMem[pOp->p1];
  if( (pIn1->flags & 0x0001)!=0 ){
    goto jump_to_p2;
  }
  break;
}

case 21: {
  int doTheJump;
  pIn1 = &aMem[pOp->p1];
  doTheJump = 3;
  if( doTheJump ) goto jump_to_p2;
  break;
}

case 51: {
  pIn1 = &aMem[pOp->p1];
  if( (pIn1->flags & 0x0001)==0 ){
    goto jump_to_p2;
  }
  break;
}

case 22: {
  if( 0 ){
    goto jump_to_p2;
  }
  break;
}

case 93: {
  u32 p2;
  u32 *aOffset;
  int len;
  int i;
  Mem *pDest;
  Mem sMem;
  const u8 *zData;
  u64 offset64;
  (void)p2;
  (void)aOffset;
  (void)len;
  (void)i;
  (void)pDest;
  (void)sMem;
  (void)zData;
  (void)offset64;

  p2 = (u32)pOp->p2;
  if( rc ) goto abort_due_to_error;
  pDest = &aMem[pOp->p3];
  if( 2 ){
    if( 0 ){
      if( 0 ){
        if( rc!=0 ) goto abort_due_to_error;
      }else{
      }
      do{
        if( 0x80 ){
        }else{
        }
      }while( 1 );

      if( 1 ){
        if( 0 ){
        }else{
          goto op_column_corrupt;
        }
      }
    }else{
    }

    if( 2 ){
      goto op_column_out;
    }
  }else{
  }

op_column_out:
  break;

op_column_corrupt:
  if( 0 ){
    pOp = &aOp[aOp[0].p3-1];
    break;
  }else{
    goto abort_due_to_error;
  }
}

case 2: {
  if( 0 ){
    if( db->flags & 0x00100000 ){
      rc = 8;
    }else{
      rc = 11;
    }
    goto abort_due_to_error;
  }
  break;
}

case 98: {
  int iMeta;
  int iDb;
  int iCookie;
  (void)iDb;
  (void)iCookie;

  iDb = pOp->p1;
  iCookie = pOp->p3;
  pOut->u.i = iMeta;
  break;
}

case 99: {
  if( rc ) goto abort_due_to_error;
  break;
}

case 100: {
  int nField;
  u32 p2;
  int iDb;
  int wrFlag;
  (void)nField;
  (void)iDb;
  (void)wrFlag;

  if( 2 ){
    goto open_cursor_set_hints;
  }

case 101:
case 112:
  if( 1 ){
    rc = (4 | (2<<8));
    goto abort_due_to_error;
  }

  nField = 0;
  p2 = (u32)pOp->p2;
  iDb = pOp->p3;
  if( pOp->opcode==112 ){
    ((void)0);
    wrFlag = 0x00000004 | (pOp->p5 & 0x08);
  }else{
    wrFlag = 0;
  }
  if( pOp->p5 & 0x10 ){
    pIn2 = &aMem[p2];
    p2 = (int)pIn2->u.i;
  }
  if( pOp->p4type==(-9) ){
    nField = 1;
  }else if( pOp->p4type==(-3) ){
    nField = pOp->p4.i;
  }
  if( 0 ) goto no_mem;

open_cursor_set_hints:
  if( rc ) goto abort_due_to_error;
  break;
}

default: {
  break;
}

    }
  }

abort_due_to_error:
  if( db->mallocFailed ){
    rc = 7;
  }
  p->rc = rc;
  if( rc==11 && db->autoCommit==0 ){
    db->flags |= ((u64)(0x00002)<<32);
  }
  rc = 1;

vdbe_return:
  while( nVmStep>=nProgressLimit && db->xProgress!=0 ){
    nProgressLimit += db->nProgressOps;
    if( db->xProgress(db->pProgressArg) ){
      nProgressLimit = (0xffffffff|(((u64)0xffffffff)<<32));
      rc = 9;
      goto abort_due_to_error;
    }
  }

  p->aCounter[4] += (int)nVmStep;
  return rc;

too_big:
  rc = 18;
  goto abort_due_to_error;

no_mem:
  rc = 7;
  goto abort_due_to_error;

abort_due_to_interrupt:
  rc = 9;
  goto abort_due_to_error;
}
