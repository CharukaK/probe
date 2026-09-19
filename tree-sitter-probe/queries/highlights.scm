; request line
(request_line method: (method) @keyword)
(request_line target: (request_target) @string.special.url)

"HTTP/" @constant.builtin
(digit) @number
"." @punctuation.delimiter

; Header fields
(field_line name: (field_name) @property)
(field_line value: (field_value) @string)
(field_value_seperator) @punctuation.delimiter

; Variable interpolation (nested inside request_target/field_value, so
; these layer specific captures on top of the @string.special.url/@string
; painted above, the same way http_version's digits/dots do)
"{{" @punctuation.special
"}}" @punctuation.special

(value_reference root: (identifier) @variable)
(member_access name: (identifier) @property)
"." @punctuation.delimiter

(index_access ["[" "]"] @punctuation.bracket)
(index_access index: (number) @number)
(index_access index: (string) @string)

(util_reference util_name: (identifier) @function.builtin)
(util_reference ["(" ")"] @punctuation.bracket)

; Multipart body markers
(multipart_field_key) @attribute
(multipart_file_key) @attribute
(field_part name: (identifier) @property)
(field_part value: (string) @string)
(file_part name: (identifier) @property)
(file_part path: (quoted_path) @string)
(file_part content_type: (string) @string)

; Raw body content
(octet_body) @string

