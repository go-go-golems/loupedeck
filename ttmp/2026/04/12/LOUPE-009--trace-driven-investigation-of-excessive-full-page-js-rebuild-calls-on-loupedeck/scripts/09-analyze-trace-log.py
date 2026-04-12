#!/usr/bin/env python3
import argparse
import re
import statistics as st
from pathlib import Path

TRACE_RE = re.compile(r'INFO (js trace|go trace) script=.* label=(\w+) seq=(\d+) event=([^ ]+) fields="([^"]*)"')


def parse_fields(s: str):
    if s in ('', 'none'):
        return {}
    out = {}
    for part in s.split(', '):
        if '=' in part:
            k, v = part.split('=', 1)
            out[k] = v
    return out


def load_events(path: Path):
    events = []
    for line in path.read_text().splitlines():
        m = TRACE_RE.search(line)
        if not m:
            continue
        source, label, seq, event, fields = m.groups()
        events.append({
            'source': source,
            'label': label,
            'seq': int(seq),
            'event': event,
            'fields': parse_fields(fields),
        })
    events.sort(key=lambda e: e['seq'])
    return events


def summarize(events):
    begins = [e for e in events if e['event'] == 'scene.renderAll.begin']
    ends = [e for e in events if e['event'] == 'scene.renderAll.end']
    loops = [e for e in events if e['event'] == 'scene.loop.tick']
    nonempty_flushes = [e for e in events if e['event'] == 'go.flush.end' and e['fields'].get('ops') == '1']

    print(f'total events: {len(events)}')
    print(f'renderAll.begin: {len(begins)}')
    print(f'renderAll.end:   {len(ends)}')
    print(f'loop.tick:       {len(loops)}')
    print(f'non-empty flushes: {len(nonempty_flushes)}')

    last_flush_seq = 0
    rebuilds_per_flush = []
    loops_per_flush = []
    for i, flush in enumerate(nonempty_flushes, 1):
        bucket = [e for e in events if last_flush_seq < e['seq'] <= flush['seq']]
        begin_count = sum(1 for e in bucket if e['event'] == 'scene.renderAll.begin')
        loop_count = sum(1 for e in bucket if e['event'] == 'scene.loop.tick')
        rebuilds_per_flush.append(begin_count)
        loops_per_flush.append(loop_count)
        print(
            f'flush {i:02d}: seq={flush["seq"]} elapsedMs={flush["fields"].get("elapsedMs", "?")} '
            f'rebuilds={begin_count} loops={loop_count}'
        )
        last_flush_seq = flush['seq']

    if rebuilds_per_flush:
        print('---')
        print(
            f'rebuilds/flush avg={sum(rebuilds_per_flush)/len(rebuilds_per_flush):.2f} '
            f'median={st.median(rebuilds_per_flush):.2f} min={min(rebuilds_per_flush)} max={max(rebuilds_per_flush)}'
        )
        print(
            f'loops/flush avg={sum(loops_per_flush)/len(loops_per_flush):.2f} '
            f'median={st.median(loops_per_flush):.2f} min={min(loops_per_flush)} max={max(loops_per_flush)}'
        )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('log', help='Path to captured trace log')
    args = parser.parse_args()
    events = load_events(Path(args.log))
    summarize(events)


if __name__ == '__main__':
    main()
