#include "tree_sitter/parser.h"
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

enum TokenType { OCTET_BODY };

static const char *BOUNDARY_MARKERS[] = {"###",   "~~~",    "@assert",
                                         "@save", "@field", "@file"};

#define BOUNDARY_COUNT (sizeof(BOUNDARY_MARKERS) / sizeof(BOUNDARY_MARKERS[0]))

static bool consume_boundary_if_present(TSLexer *lexer) {
  bool alive[BOUNDARY_COUNT];
  for (size_t i = 0; i < BOUNDARY_COUNT; i++)
    alive[i] = true;

  size_t depth = 0;
  for (;;) {
    bool any_need_more = false;
    for (size_t i = 0; i < BOUNDARY_COUNT; i++) {
      if (alive[i] && BOUNDARY_MARKERS[i][depth] != '\0')
        any_need_more = true;
    }

    if (!any_need_more)
      break;

    int32_t c = lexer->lookahead;
    bool any_alive = false;
    for (size_t i = 0; i < BOUNDARY_COUNT; i++) {
      if (!alive[i])
        continue;
      char expected = BOUNDARY_MARKERS[i][depth];
      if (expected == '\0' || (int32_t)(unsigned char)expected != c)
        alive[i] = false;
      else
        any_alive = true;
    }
    if (!any_alive)
      return false;
    lexer->advance(lexer, false);
    depth++;
  }

  int32_t next = lexer->lookahead;
  return next == ' ' || next == '\t' || next == '\r' || next == '\n' ||
         lexer->eof(lexer);
}

void *tree_sitter_probe_external_scanner_create(void) { return NULL; }
void tree_sitter_probe_external_scanner_destroy(void *p) { (void)p; }
unsigned tree_sitter_probe_external_scanner_serialize(void *p, char *b) { (void)p;(void)b; return 0; }
void tree_sitter_probe_external_scanner_deserialize(void *p, const char *b, unsigned n) { (void)p;(void)b;(void)n; }

bool tree_sitter_probe_external_scanner_scan(void *payload, TSLexer *lexer, const bool *valid_symbols) {
  (void)payload;
  if (!valid_symbols[OCTET_BODY]) return false;

  bool at_line_start = true, consumed_any = false;
  for (;;) {
    lexer->mark_end(lexer);              // commit everything confirmed-safe
    if (lexer->eof(lexer)) break;
    if (at_line_start) {
      if (consume_boundary_if_present(lexer)) break;   // real boundary — stop, exclude it
      at_line_start = false;
      continue;                          // fold whatever was speculatively consumed into the token
    }
    int32_t c = lexer->lookahead;
    lexer->advance(lexer, false);
    consumed_any = true;
    if (c == '\n') at_line_start = true;
  }

  if (!consumed_any) return false;       // body started right at a boundary — decline entirely
  lexer->result_symbol = OCTET_BODY;
  return true;
}

