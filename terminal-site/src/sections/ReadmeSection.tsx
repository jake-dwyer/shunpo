import { Chevron } from '../components/Chevron'
import { CodeList, CodeLine, Comment } from '../components/Code'

export function ReadmeSection() {
  return (
    <section id="readme" aria-label="About shunpo" className="max-w-[1100px]">
      <CodeList>
        <CodeLine>
          <Comment>
            # a typing test for your terminal,
            <br />
            built for the moments between builds,
            <br />
            deploys, and whatever your agent is doing.
          </Comment>
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment># built with go + bubble tea</Comment>
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <a href="https://github.com/jake-dwyer/shunpo" className="text-inherit no-underline [text-underline-offset:6px]">
            <strong>
              shunpo <span className="text-[#73787b]">/ 瞬歩</span>
            </strong>
            <Chevron />
          </a>
        </CodeLine>
      </CodeList>
    </section>
  )
}
