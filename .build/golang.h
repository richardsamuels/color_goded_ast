struct Callback {

};
void Errored(struct Callback c, char* msg);
void InsertHighlight(struct Callback c, char * type, int line, int column, char* token);
