grammar Azm;

relation
    :   rel ('|' rel)*  EOF
    ;

permission
    :   union EOF           #UnionPerm
    |   intersection EOF    #IntersectionPerm
    |   exclusion EOF       #ExclusionPerm
    ;

union
    :   perm ('|' perm)*
    ;

intersection
    :   perm '&' perm ('&' perm)*
    ;

exclusion
    :   perm '-' perm
    ;

rel
    :   ID                  #DirectRel
    |   ID COLON ASTERISK   #WildcardRel
    |   ID HASH ID          #SubjectRel
    ;

perm
    :   ID            #DirectPerm
    |   ID ARROW ID   #ArrowPerm
    ;

ARROW:
    '->' ;

HASH:
    '#' ;

COLON:
    ':' ;

ASTERISK:
    '*' ;

ID: [a-zA-Z][a-zA-Z0-9._-]*[a-zA-Z0-9] ;

WS: [ \t\n\r\f]+ -> skip ;

ERROR: . ;
