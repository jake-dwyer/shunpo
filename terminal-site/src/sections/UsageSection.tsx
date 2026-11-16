import { Annotation, CodeList, CodeLine, Comment, Kbd } from '../components/Code'

export function UsageSection() {
  return (
    <section id="usage" aria-label="Usage" className="max-w-[1100px]">
      <CodeList>
        <CodeLine>
          <code className="font-[inherit]">shunpo</code>
          <Annotation># 30 seconds on your first run</Annotation>
        </CodeLine>
        <CodeLine>
          <code className="font-[inherit]">shunpo -time 60</code>
          <Annotation># take a minute</Annotation>
        </CodeLine>
        <CodeLine>
          <code className="font-[inherit]">shunpo -words 50</code>
          <Annotation># finish 50 words</Annotation>
        </CodeLine>
        <CodeLine>
          <code className="font-[inherit]">shunpo -theme dracula</code>
          <Annotation># set the mood</Annotation>
        </CodeLine>
        <CodeLine blank />
        <CodeLine>
          <Comment># your keyboard is all you need</Comment>
        </CodeLine>
        <CodeLine>
          <Kbd>tab</Kbd>
          <Annotation>settings · restart from results</Annotation>
        </CodeLine>
        <CodeLine>
          <Kbd>enter</Kbd>
          <Annotation>fresh words, same settings</Annotation>
        </CodeLine>
        <CodeLine>
          <Kbd>esc</Kbd>
          <Annotation>quit · close settings when open</Annotation>
        </CodeLine>
      </CodeList>
    </section>
  )
}
