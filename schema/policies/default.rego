package dawg.sanitizer

import rego.v1

default allow := false

allow if {
  input.fieldsRedacted >= 0
  count(input.blockedFields) == 0
}