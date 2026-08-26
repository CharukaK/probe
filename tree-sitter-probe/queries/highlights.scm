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
(path_accessor) @punctuation.delimiter

(util_reference util_name: (identifier) @function.builtin)
(util_reference "(" @punctuation.bracket)
(util_reference ")" @punctuation.bracket)

; Multipart body markers
"@field" @attribute
"@file" @attribute

; Raw body content
(octet_body) @string

