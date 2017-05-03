module Basics exposing
  ( (+), (-)
  )

import Native.Basics

(+) : number -> number -> number
(+) = 
    Native.Basics.add

(-) : number -> number -> number
(-) = 
    Native.Basics.add