module List exposing (..)

import Native.List

(::) : a -> List a -> List a
(::) =
  Native.List.cons