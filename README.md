# Integration HTTP health probe container (Go) with Azure Application Insights custom availability test

Track public/private site availability

## Overview

<img src="https://raw.githubusercontent.com/ToruMakabe/Images/master/healthprobe.jpg?raw=true" width="500">

## Typical usage

This probe does not depend much on the platform because using Go and Docker. But the following is a typical usage.

* test probe app (run CI action when push any branches)
* prepare infrastructure (apply under /terraform dir)
* build & push probe app container (run build action when push tag with semantic versioning vX.Y.Z)
* list target site to csv file (default: sample_target_mnt.csv under /conf dir. You can change filename/path in deploy action script)
* deploy probe app container (run deploy action manually)
* cleanup probe app container (run cleanup action manually)
* cleanup infrastructure (destroy under /terraform dir)
