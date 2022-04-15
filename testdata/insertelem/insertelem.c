typedef struct Hash Hash;
typedef struct HashElem HashElem;

struct Hash {
  unsigned int htsize;
  unsigned int count;
  HashElem *first;
  struct _ht {
    unsigned int count;
    HashElem *chain;
  } *ht;
};

struct HashElem {
  HashElem *next, *prev;
  void *data;
  const char *pKey;
};

static void insertElement(
  Hash *pH,
  struct _ht *pEntry,
  HashElem *pNew
){
  HashElem *pHead;
  if( pEntry ){
    pHead = pEntry->count ? pEntry->chain : 0;
    pEntry->count++;
    pEntry->chain = pNew;
  }else{
    pHead = 0;
  }
  if( pHead ){
    pNew->next = pHead;
    pNew->prev = pHead->prev;
    if( pHead->prev ){ pHead->prev->next = pNew; }
    else { pH->first = pNew; }
    pHead->prev = pNew;
  }else{
    pNew->next = pH->first;
    if( pH->first ){ pH->first->prev = pNew; }
    pNew->prev = 0;
    pH->first = pNew;
  }
}

int main() {
    return 0;
}
