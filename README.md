# [ZikiChombo](http://zikichombo.org) codec project

[![Build Status](https://travis-ci.com/zikichombo/codec.svg?branch=master)](https://travis-ci.com/zikichombo/codec)

# Status:

## Containers
* [-] ogg
* [-] caf
* [-] webm


## Codecs

Below we list some common codecs and their status.  Note that
darwin/iOS may use system codec support independently of whether or
not go implementations are available.  Some codecs may have limitations
in their degree of support.  We consider a codec supported in the 
table below if common use cases are supported.  We exclude codecs for
which we know of active patent liabilities, such as aac.  Note that
some such codecs may have encoders and decoders licensed with the 
host OS (such as darwin/iOS).


| Codec | source | sink | source+seek | random-access | registered |
|-------|--------|------|-------------|---------------|------------|
| wav   | +      | +    | +           | -             | -          |
| flac  | +      | -    | -           | -             | -          |
| opus  | -      | -    | -           | -             | -          |
| vorbis | -     | -    | -           | -             | -          |
| aif   | -      | -    | -           | -             | -          |
| speex | -      | -    | -           | -             | -          |
| mp3   | -      | -    | -           | -             | -          |

## Ext codecs
The following are the codecs implemented in zikichombo.org/ext due to import direction
impasse between developers.  Others are found here.

* flac

## 3rd party implementations
Below lists the implementations of codecs which interoperate with zikichombo and
are hosted by 3rd parties.

* example, gitlab.com/example-zc-codec







