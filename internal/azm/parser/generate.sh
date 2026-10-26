#!/bin/sh

# alias antlr4='java -Xmx500M -cp "./antlr-4.13.1-complete.jar:$CLASSPATH" org.antlr.v4.Tool'
antlr -Dlanguage=Go -visitor -no-listener -package parser *.g4
