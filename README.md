# `weakand`

[![Build Status](https://travis-ci.org/wangkuiyi/weakand.svg?branch=develop)](https://travis-ci.org/wangkuiyi/weakand)

`weakand` is a Go implementation of the Weak-AND retrieval algorithm
and a search engine.  The search engine includes an HTTP frontend
server and a Go RPC backend server, where the backend server maintains
the inverted index and runs the Weak-AND retrieval algorithm.
