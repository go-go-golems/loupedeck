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
        events.append({'source': source, 'label': label, 'seq': int(seq), 'event': event, 'fields': parse_fields(fields)})
    events.sort(key=lambda e: e['seq'])
    return events

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('log')
    args = parser.parse_args()
    events = load_events(Path(args.log))
    begins = [e for e in events if e['event'] == 'scene.renderAll.begin']
    loops = [e for e in events if e['event'] == 'scene.loop.tick']
    flushes = [e for e in events if e['event'] == 'go.flush.end' and e['fields'].get('ops') == '1']
    print(f'total events: {len(events)}')
    print(f'renderAll.begin: {len(begins)}')
    print(f'loop.tick: {len(loops)}')
    print(f'non-empty flushes: {len(flushes)}')
    last_seq = 0
    rebuilds = []
    for i, flush in enumerate(flushes, 1):
        bucket = [e for e in events if last_seq < e['seq'] <= flush['seq']]
        n = sum(1 for e in bucket if e['event'] == 'scene.renderAll.begin')
        l = sum(1 for e in bucket if e['event'] == 'scene.loop.tick')
        rebuilds.append(n)
        print(f'flush {i:02d}: seq={flush["seq"]} elapsedMs={flush["fields"].get("elapsedMs", "?")} rebuilds={n} loops={l}')
        last_seq = flush['seq']
    if rebuilds:
        print('---')
        print(f'rebuilds/flush avg={sum(rebuilds)/len(rebuilds):.2f} median={st.median(rebuilds):.2f} min={min(rebuilds)} max={max(rebuilds)}')

if __name__ == '__main__':
    main()
