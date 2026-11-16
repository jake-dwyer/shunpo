import { CodeList, CodeLine, Comment, Command, CopyCommand } from '../components/Code'
import { MiniDemo } from '../components/MiniDemo'

const SPEEDLINES = `                               ╱    ╱     ╱    ╱    ╱  ╱       ╱╱  ╱    ╱ ╱
                              ╱    ╱     ╱    ╱    ╱  ╱       ╱╱  ╱    ╱
                             ╱    ╱     ╱    ╱ ╱  ╱  ╱        ╱  ╱    ╱
                            ╱    ╱     ╱    ╱ ╱  ╱  ╱        ╱  ╱    ╱
                         ╱ ╱    ╱     ╱    ╱ ╱  ╱    ╱      ╱  ╱    ╱    ╱
                        ╱ ╱    ╱     ╱    ╱ ╱  ╱    ╱      ╱  ╱         ╱
                       ╱ ╱    ╱     ╱    ╱ ╱  ╱    ╱      ╱  ╱         ╱
                      ╱ ╱    ╱     ╱    ╱ ╱       ╱      ╱  ╱         ╱
                     ╱ ╱    ╱     ╱ ╱  ╱ ╱       ╱      ╱            ╱
                    ╱ ╱    ╱     ╱ ╱  ╱ ╱       ╱      ╱            ╱
                   ╱ ╱    ╱     ╱ ╱    ╱       ╱      ╱            ╱
                  ╱ ╱    ╱     ╱ ╱    ╱       ╱        ╱          ╱
             ╱   ╱ ╱    ╱     ╱ ╱    ╱       ╱        ╱
            ╱   ╱ ╱    ╱ ╱     ╱    ╱       ╱        ╱
           ╱   ╱ ╱    ╱ ╱     ╱    ╱                ╱
          ╱   ╱ ╱      ╱     ╱    ╱                ╱
         ╱   ╱ ╱      ╱     ╱                     ╱
        ╱   ╱ ╱      ╱     ╱
       ╱   ╱        ╱     ╱
      ╱   ╱        ╱
     ╱            ╱
    ╱`

export function TerminalSection() {
  return (
    <section id="terminal" aria-label="Install shunpo" className="relative max-w-[1100px]">
      <pre
        aria-hidden="true"
        className="pointer-events-none absolute -top-[40px] -right-[20px] z-0 m-0 text-[clamp(14px,1.3vw,19px)] leading-[1.3] text-accent opacity-[0.07] select-none max-[900px]:hidden"
      >
        {SPEEDLINES}
      </pre>
      <CodeList className="relative z-10">
        <CodeLine>
          <Comment># flash step. a typing test for your terminal.</Comment>
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment>
            # for the moments between builds, deploys,
            <br />
            and whatever your agent is doing.
          </Comment>
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment># install with go</Comment>
        </CodeLine>
        <CodeLine>
          <CopyCommand command="go install github.com/jake-dwyer/shunpo@latest" />
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment># open a spare pane. make a little progress.</Comment>
        </CodeLine>
        <CodeLine>
          <Command command="shunpo" />
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment># a quick look</Comment>
        </CodeLine>
        <CodeLine>
          <MiniDemo />
        </CodeLine>
      </CodeList>
    </section>
  )
}
