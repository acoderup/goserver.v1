package logger

var data = `<seelog type="adaptive" mininterval="2000000" maxinterval="100000000" critmsgcount="500" minlevel="trace">
        <exceptions>
            <exception filepattern="test*" minlevel="error"/>
        </exceptions>
        <outputs formatid="all">
            <rollingfile formatid="all" type="size" filename="./all.log" maxsize="50000000" maxrolls="5" />
            <filter levels="info,trace,warn">
              <console formatid="fmtinfo"/>
            </filter>
            <filter levels="error,critical" formatid="fmterror">
              <console/>
              <file path="errors.log"/>
            </filter>
        </outputs>
        <formats>
            <format id="fmtinfo" format="[%Date][%Time] [%Level] %Msg%n"/>
            <format id="fmterror" format="[%Date][%Time] [%LEVEL] [%FuncShort @ %File.%Line] %Msg%n"/>
            <format id="all" format="[%Date][%Time] [%Level] [@ %File.%Line] %Msg%n"/>
            <format id="criticalemail" format="Critical error on our server!\n    %Time %Date %RelFile %Func %Msg \nSent by Seelog"/>
        </formats>
    </seelog>`
