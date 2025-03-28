package viperx

// common executor timer
var data = `{
  "module": {
    "Options": {
      "QueueBacklog": 1024,
      "MaxDone": 1024,
      "Interval": 100
    }
  },

  "executor": {
    "Options": {
      "QueueBacklog": 1024,
      "MaxDone": 1024,
      "Interval": 0
    },
    "Worker": {
      "WorkerCnt": 8,
      "Options": {
        "QueueBacklog": 1024,
        "MaxDone": 1024,
        "Interval": 0
      }
    }
  },

  "timer": {
    "Options": {
      "QueueBacklog": 1024,
      "MaxDone": 1024,
      "Interval": 100
    }
  },

  "core": {
    "MaxProcs": 4
  },

  "cmdline": {
    "SupportCmdLine": true
  },
  "signal":{
    "SupportSignal": true
  },
  "profile" : {
    "SlowMS": 500
  }
}`
