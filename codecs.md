# ZikiChombo Codec Developer Notes

This document summarizes the major components of the 
design discussions that took place from the issue tracker.
Many thanks to all that particpated.

## Don't write codecs.
The first an most important point is that to work on codecs
in a way that can interoperate with other zc components 
only requires implementing whatever subset of the 

    sound.{Sink,Source,SourceSeeker,RandomAccess}.

interfaces that is of interest.  If you are not competing to make the next
greatest Go implementation of an established codec, or simply the next greatest
codec, you can stop here.  It's that simple.

## Codec plugins
Codec plugins enable codec providers to register codecs 
and codec consumers to coordinate what happens when 
an application requests that sound be encoded or decoded.

This coordination occurs based on package path of the 
implementing codec.  As a result all codecs are associated
with a package path.  Since Go only uses acyclic package 
dependencies, this leads us to the following constraint.

```
Either zikichombo.org/codec imports your codec, or you
import zikichombo.org/codec, but not both.
```

### zc imports your codec
If zikichombo.org imports your codec, we need to coordinate
some things to ensure our consumers get what they expect 
when they use zikichombo.  We are interested in importing 
the best codecs we can given our constraints.  

### you import zc
If you import zikichombo.org, our rules are simple:
semver versions, go modules, standard BSD3 clause license.
We do not have resources to guarantee release schedules at 
the time of this writing.

It is possible to wrap the essence of a codec in two packages
and have imports going both directions, but the package name
and thus what is formally in ZikiChombo is "the codec", would 
differ.


