import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";

type RawRecord = Record<string, unknown>;

type LogSource = {
  name: string;
  text: string;
};

type DefaultLogsResponse = {
  sources?: LogSource[];
};

type ParsedEvent = {
  id: string;
  fileName: string;
  lineNumber: number;
  ordinal: number;
  rawLine: string;
  raw: RawRecord;
  parseError: boolean;
  source: string;
  level: string;
  message: string;
  timeText: string;
  timeMs: number;
  hasTime: boolean;
  fields: LogField[];
  searchText: string;
};

type LogField = {
  key: string;
  value: string;
};

type FieldOption = {
  key: string;
  count: number;
};

type FieldIndex = {
  fields: FieldOption[];
  eventsByFieldValue: Map<string, Map<string, ParsedEvent[]>>;
};

type GcfCount = {
  value: string;
  count: number;
};

type SortDirection = "asc" | "desc";

type FileSystemFileEntry = {
  kind: "file";
  name: string;
  getFile: () => Promise<File>;
};

type FileSystemDirectoryEntry = {
  kind: "directory";
  values: () => AsyncIterable<FileSystemFileEntry | FileSystemDirectoryEntry>;
};

type DirectoryPickerWindow = Window & {
  showDirectoryPicker?: () => Promise<FileSystemDirectoryEntry>;
};

const TIME_KEYS = ["time", "timestamp", "ts", "@timestamp", "datetime", "date"];
const MESSAGE_KEYS = ["msg", "message", "event", "error"];
const LEVEL_KEYS = ["level", "severity"];
const SOURCE_KEYS = ["source", "log_source", "logger", "service", "side"];
const FIELD_EXCLUDE_KEYS = new Set([
  ...TIME_KEYS,
  ...MESSAGE_KEYS,
  ...LEVEL_KEYS,
  ...SOURCE_KEYS,
]);
const LEVEL_ORDER = ["TRACE", "DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL"];
const SOURCE_ORDER = ["client", "server", "app"];
const PAGE_SIZES = [50, 100, 200, 300, 500, 1000];
const DEFAULT_PAGE_SIZE = 100;
const DEFAULT_CONTEXT_RADIUS = 20;
const LOG_FILE_PATTERN = /(^.*\.(log|jsonl|json|txt)$)/i;
const DEFAULT_LOGS_ENDPOINT = "/default-logs";

const numberFormatter = new Intl.NumberFormat();

function App() {
  const [events, setEvents] = useState<ParsedEvent[]>([]);
  const [status, setStatus] = useState("No logs loaded");
  const [statusIsError, setStatusIsError] = useState(false);
  const [selectedSources, setSelectedSources] = useState<Set<string>>(new Set());
  const [selectedLevels, setSelectedLevels] = useState<Set<string>>(new Set());
  const [selectedFieldKey, setSelectedFieldKey] = useState("");
  const [fieldFilterValue, setFieldFilterValue] = useState("");
  const [selectedGcf, setSelectedGcf] = useState<string | null>(null);
  const [searchText, setSearchText] = useState("");
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [contextRadiusText, setContextRadiusText] = useState(String(DEFAULT_CONTEXT_RADIUS));
  const [isContextViewActive, setIsContextViewActive] = useState(false);
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [currentPage, setCurrentPage] = useState(1);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const defaultLogsRequestedRef = useRef(false);

  useEffect(() => {
    if (defaultLogsRequestedRef.current) {
      return;
    }
    defaultLogsRequestedRef.current = true;
    void loadDefaultLogs();
  }, []);

  const sourceCounts = useMemo(() => {
    const counts = countBy(events, (event) => event.source);
    return Array.from(counts.entries()).sort((a, b) => {
      const sourceDelta = sourceRank(a[0]) - sourceRank(b[0]);
      return sourceDelta || a[0].localeCompare(b[0]);
    });
  }, [events]);

  const levelCounts = useMemo(() => {
    const counts = countBy(events, (event) => event.level);
    return Array.from(counts.entries()).sort((a, b) => {
      const levelDelta = levelRank(a[0]) - levelRank(b[0]);
      return levelDelta || a[0].localeCompare(b[0]);
    });
  }, [events]);

  const textFilteredEvents = useMemo(() => {
    const terms = parseSearchTerms(searchText);

    return events.filter((event) => {
      if (!selectedSources.has(event.source)) {
        return false;
      }
      if (!selectedLevels.has(event.level)) {
        return false;
      }

      for (const term of terms) {
        if (!event.searchText.includes(term)) {
          return false;
        }
      }

      return true;
    });
  }, [events, searchText, selectedLevels, selectedSources]);

  const gcfCounts = useMemo(() => topGcfCounts(textFilteredEvents, 5), [textFilteredEvents]);

  const gcfFilteredEvents = useMemo(() => {
    if (selectedGcf === null) {
      return textFilteredEvents;
    }

    return textFilteredEvents.filter((event) => eventHasFieldValue(event, "gcf", selectedGcf));
  }, [selectedGcf, textFilteredEvents]);

  const fieldIndex = useMemo(() => buildFieldIndex(gcfFilteredEvents), [gcfFilteredEvents]);

  const baseFilteredEvents = useMemo(
    () => applyFieldValueFilter(gcfFilteredEvents, fieldIndex, selectedFieldKey, fieldFilterValue),
    [fieldFilterValue, fieldIndex, gcfFilteredEvents, selectedFieldKey],
  );

  const isFieldFilterActive = selectedFieldKey !== "" && fieldFilterValue.trim() !== "";

  const filteredEvents = useMemo(() => {
    const filtered = [...baseFilteredEvents];
    sortEvents(filtered, sortDirection);
    return filtered;
  }, [baseFilteredEvents, sortDirection]);

  const unfilteredSortedEvents = useMemo(() => {
    const sorted = [...events];
    sortEvents(sorted, sortDirection);
    return sorted;
  }, [events, sortDirection]);

  const selectedEvent = useMemo(
    () => events.find((event) => event.id === selectedEventId) || null,
    [events, selectedEventId],
  );
  const contextRadius = parseContextRadius(contextRadiusText);
  const contextEvents = useMemo(() => {
    if (!isContextViewActive || selectedEventId === null || contextRadius === null) {
      return [];
    }

    const selectedIndex = unfilteredSortedEvents.findIndex((event) => event.id === selectedEventId);
    if (selectedIndex === -1) {
      return [];
    }

    const start = Math.max(0, selectedIndex - contextRadius);
    const end = Math.min(unfilteredSortedEvents.length, selectedIndex + contextRadius + 1);
    return unfilteredSortedEvents.slice(start, end);
  }, [contextRadius, isContextViewActive, selectedEventId, unfilteredSortedEvents]);
  const timelineEvents = isContextViewActive ? contextEvents : filteredEvents;

  const pageCount = isContextViewActive ? 1 : Math.max(1, Math.ceil(timelineEvents.length / pageSize));
  const safeCurrentPage = Math.min(currentPage, pageCount);
  const pageStart = timelineEvents.length && !isContextViewActive ? (safeCurrentPage - 1) * pageSize : 0;
  const visibleEvents = isContextViewActive ? timelineEvents : timelineEvents.slice(pageStart, pageStart + pageSize);
  const pageEnd = pageStart + visibleEvents.length;

  useEffect(() => {
    setCurrentPage((page) => Math.min(page, pageCount));
  }, [pageCount]);

  useEffect(() => {
    if (selectedEventId !== null && !selectedEvent) {
      setSelectedEventId(null);
      setIsContextViewActive(false);
    }
  }, [selectedEvent, selectedEventId]);

  async function loadDefaultLogs() {
    setStatusMessage("Loading default client/server logs...");

    try {
      const response = await fetch(DEFAULT_LOGS_ENDPOINT, { cache: "no-store" });
      if (!response.ok) {
        setStatusMessage("Default client/server logs were not found. Use Open logs instead.", true);
        return;
      }

      const payload = (await response.json()) as DefaultLogsResponse;
      const sources = (payload.sources || []).filter((source) => source.name && source.text);
      if (!sources.length) {
        setStatusMessage("Default client/server logs were empty. Use Open logs instead.", true);
        return;
      }

      replaceEvents(parseSources(sources));
      setStatusMessage(statusForLoadedFiles(sources));
    } catch (error) {
      setStatusMessage(`Failed to load default logs: ${getErrorMessage(error)}`, true);
    }
  }

  async function loadFiles(files: File[]) {
    if (!files.length) {
      return;
    }

    setStatusMessage(`Reading ${files.length} file${files.length === 1 ? "" : "s"}...`);

    try {
      const sources = await Promise.all(
        files.map(async (file) => ({
          name: file.name,
          text: await file.text(),
        })),
      );

      replaceEvents(parseSources(sources));
      setStatusMessage(statusForLoadedFiles(sources));
    } catch (error) {
      setStatusMessage(`Failed to read logs: ${getErrorMessage(error)}`, true);
    }
  }

  async function openLogFolder() {
    const pickerWindow = window as DirectoryPickerWindow;
    if (!pickerWindow.showDirectoryPicker) {
      setStatusMessage("Folder picker is not supported by this browser. Use Open logs instead.", true);
      return;
    }

    try {
      const directory = await pickerWindow.showDirectoryPicker();
      const files = await collectLogFiles(directory);
      if (!files.length) {
        setStatusMessage("No log files found in the selected folder.", true);
        return;
      }
      await loadFiles(files);
    } catch (error) {
      if (isAbortError(error)) {
        return;
      }
      setStatusMessage(`Failed to open folder: ${getErrorMessage(error)}`, true);
    }
  }

  function replaceEvents(nextEvents: ParsedEvent[]) {
    setEvents(nextEvents);
    setSelectedSources(new Set(uniqueValues(nextEvents, "source")));
    setSelectedLevels(new Set(uniqueValues(nextEvents, "level")));
    setSelectedFieldKey("");
    setFieldFilterValue("");
    setSelectedGcf(null);
    setSelectedEventId(null);
    setIsContextViewActive(false);
    setContextRadiusText(String(DEFAULT_CONTEXT_RADIUS));
    resetPagination();
  }

  function clearLogs() {
    setEvents([]);
    setSelectedSources(new Set());
    setSelectedLevels(new Set());
    setSelectedFieldKey("");
    setFieldFilterValue("");
    setSelectedGcf(null);
    setSearchText("");
    setSelectedEventId(null);
    setIsContextViewActive(false);
    setContextRadiusText(String(DEFAULT_CONTEXT_RADIUS));
    resetPagination();
    setStatusMessage("No logs loaded");
  }

  function setStatusMessage(message: string, isError = false) {
    setStatus(message);
    setStatusIsError(isError);
  }

  function onFileInputChange(event: ChangeEvent<HTMLInputElement>) {
    void loadFiles(Array.from(event.currentTarget.files || []));
    event.currentTarget.value = "";
  }

  function toggleSource(source: string) {
    setSelectedSources((current) => toggleSetValue(current, source));
    resetPagination();
  }

  function toggleLevel(level: string) {
    setSelectedLevels((current) => toggleSetValue(current, level));
    resetPagination();
  }

  function toggleGcf(gcf: string) {
    setSelectedGcf((current) => (current === gcf ? null : gcf));
    resetPagination();
  }

  function selectEvent(eventId: string) {
    setSelectedEventId(eventId);
    resetPagination();
  }

  function resetPagination() {
    setCurrentPage(1);
  }

  function downloadFiltered() {
    if (!filteredEvents.length) {
      return;
    }

    const content = `${filteredEvents.map((event) => event.rawLine).join("\n")}\n`;
    const blob = new Blob([content], { type: "application/x-ndjson" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "filtered-logs.jsonl";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="appShell">
      <header className="topBar">
        <div className="titleBlock">
          <h1>Log Timeline</h1>
          <div className={`statusText ${statusIsError ? "error" : ""}`}>{status}</div>
        </div>
        <div className="topActions">
          <button type="button" className="primaryButton" onClick={() => fileInputRef.current?.click()}>
            Open logs
          </button>
          <button type="button" onClick={() => void openLogFolder()}>
            Open folder
          </button>
          <button type="button" disabled={!filteredEvents.length} onClick={downloadFiltered}>
            Download filtered
          </button>
          <button type="button" className="quietButton" onClick={clearLogs}>
            Clear
          </button>
          <input
            ref={fileInputRef}
            className="hiddenInput"
            type="file"
            multiple
            accept=".log,.json,.jsonl,.txt,application/json"
            onChange={onFileInputChange}
          />
        </div>
      </header>

      <main className="workspace">
        <aside className="filterPanel" aria-label="Filters">
          <section className="filterGroup">
            <label className="fieldLabel" htmlFor="searchInput">
              Search
            </label>
            <input
              id="searchInput"
              type="search"
              placeholder="message, field, source"
              autoComplete="off"
              value={searchText}
              onChange={(event) => {
                setSearchText(event.target.value);
                resetPagination();
              }}
            />
          </section>

          <ContextControls
            selectedEvent={selectedEvent}
            contextRadiusText={contextRadiusText}
            isContextViewActive={isContextViewActive}
            hasInvalidContextRadius={contextRadius === null}
            onContextRadiusChange={setContextRadiusText}
            onToggleContext={() => {
              setIsContextViewActive((active) => {
                if (active) {
                  setSelectedEventId(null);
                }
                return !active;
              });
              resetPagination();
            }}
          />

          <GcfFilter
            items={gcfCounts}
            selectedGcf={selectedGcf}
            onToggle={toggleGcf}
            onClear={() => {
              setSelectedGcf(null);
              resetPagination();
            }}
          />

          <FieldValueFilter
            fields={fieldIndex.fields}
            selectedFieldKey={selectedFieldKey}
            fieldFilterValue={fieldFilterValue}
            isActive={isFieldFilterActive}
            onFieldChange={(fieldKey) => {
              setSelectedFieldKey(fieldKey);
              setFieldFilterValue("");
              resetPagination();
            }}
            onValueChange={(value) => {
              setFieldFilterValue(value);
              resetPagination();
            }}
            onClear={() => {
              setSelectedFieldKey("");
              setFieldFilterValue("");
              resetPagination();
            }}
          />

          <FilterGroup
            title="Sources"
            items={sourceCounts}
            selected={selectedSources}
            onToggle={toggleSource}
            onAll={() => {
              setSelectedSources(new Set(uniqueValues(events, "source")));
              resetPagination();
            }}
            onNone={() => {
              setSelectedSources(new Set());
              resetPagination();
            }}
          />

          <FilterGroup
            title="Levels"
            items={levelCounts}
            selected={selectedLevels}
            onToggle={toggleLevel}
            onAll={() => {
              setSelectedLevels(new Set(uniqueValues(events, "level")));
              resetPagination();
            }}
            onNone={() => {
              setSelectedLevels(new Set());
              resetPagination();
            }}
          />
        </aside>

        <section className="timelinePanel" aria-label="Timeline">
          <div className="summaryGrid" aria-label="Log summary">
            <Metric label="Loaded" value={formatNumber(events.length)} />
            <Metric label="Visible" value={formatNumber(timelineEvents.length)} />
          </div>

          <div className="timelineToolbar">
            <div className="resultMeta">
              {isContextViewActive
                ? contextResultMeta(events, selectedEvent, pageStart, pageEnd, timelineEvents.length)
                : resultMeta(events, pageStart, pageEnd, timelineEvents.length)}
            </div>
            <div className="timelineControls">
              <label className="sortControl">
                Sort
                <select
                  value={sortDirection}
                  onChange={(event) => {
                    setSortDirection(event.target.value as SortDirection);
                    resetPagination();
                  }}
                >
                  <option value="asc">Oldest first</option>
                  <option value="desc">Newest first</option>
                </select>
              </label>
              <label className="sortControl">
                Page size
                <select
                  value={pageSize}
                  onChange={(event) => {
                    setPageSize(Number(event.target.value));
                    resetPagination();
                  }}
                >
                  {PAGE_SIZES.map((size) => (
                    <option value={size} key={size}>
                      {size}
                    </option>
                  ))}
                </select>
              </label>
              <PaginationControls
                currentPage={safeCurrentPage}
                pageCount={pageCount}
                disabled={isContextViewActive}
                onPageChange={setCurrentPage}
              />
            </div>
          </div>

          <div className="timeline">
            {visibleEvents.length ? (
              visibleEvents.map((event, index) => (
                <EventRow
                  key={event.id}
                  event={event}
                  previousEvent={pageStart + index > 0 ? timelineEvents[pageStart + index - 1] : undefined}
                  selected={selectedEventId === event.id}
                  onSelect={() => selectEvent(event.id)}
                />
              ))
            ) : (
              <div className="emptyTimeline">
                {events.length
                  ? isContextViewActive
                    ? "No context events"
                    : "No events match the current filters"
                  : "Open logs or open a folder"}
              </div>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}

function PaginationControls({
  currentPage,
  pageCount,
  disabled,
  onPageChange,
}: {
  currentPage: number;
  pageCount: number;
  disabled?: boolean;
  onPageChange: (page: number) => void;
}) {
  return (
    <nav className="paginationControls" aria-label="Pagination">
      <button type="button" disabled={disabled || currentPage <= 1} onClick={() => onPageChange(1)}>
        First
      </button>
      <button type="button" disabled={disabled || currentPage <= 1} onClick={() => onPageChange(currentPage - 1)}>
        Previous
      </button>
      <span>{`Page ${formatNumber(currentPage)} of ${formatNumber(pageCount)}`}</span>
      <button
        type="button"
        disabled={disabled || currentPage >= pageCount}
        onClick={() => onPageChange(currentPage + 1)}
      >
        Next
      </button>
      <button type="button" disabled={disabled || currentPage >= pageCount} onClick={() => onPageChange(pageCount)}>
        Last
      </button>
    </nav>
  );
}

function ContextControls({
  selectedEvent,
  contextRadiusText,
  isContextViewActive,
  hasInvalidContextRadius,
  onContextRadiusChange,
  onToggleContext,
}: {
  selectedEvent: ParsedEvent | null;
  contextRadiusText: string;
  isContextViewActive: boolean;
  hasInvalidContextRadius: boolean;
  onContextRadiusChange: (value: string) => void;
  onToggleContext: () => void;
}) {
  return (
    <section className="filterGroup">
      <div className="groupHeader">
        <h2>Line Context</h2>
      </div>
      <div className="contextControls">
        <label>
          <span>Before/after</span>
          <input
            type="number"
            min="0"
            step="1"
            inputMode="numeric"
            value={contextRadiusText}
            aria-invalid={hasInvalidContextRadius}
            onChange={(event) => onContextRadiusChange(event.target.value)}
          />
        </label>
        <button
          type="button"
          className="contextToggleButton"
          aria-pressed={isContextViewActive}
          disabled={!isContextViewActive && (selectedEvent === null || hasInvalidContextRadius)}
          onClick={onToggleContext}
        >
          Show context
        </button>
      </div>
    </section>
  );
}

function FieldValueFilter({
  fields,
  selectedFieldKey,
  fieldFilterValue,
  isActive,
  onFieldChange,
  onValueChange,
  onClear,
}: {
  fields: FieldOption[];
  selectedFieldKey: string;
  fieldFilterValue: string;
  isActive: boolean;
  onFieldChange: (fieldKey: string) => void;
  onValueChange: (value: string) => void;
  onClear: () => void;
}) {
  return (
    <section className="filterGroup">
      <div className="groupHeader">
        <h2>Field Value</h2>
        <button type="button" disabled={!selectedFieldKey && !fieldFilterValue} onClick={onClear}>
          Clear
        </button>
      </div>
      <div className="fieldValueControls">
        <label>
          <span>Field</span>
          <select
            value={selectedFieldKey}
            disabled={!fields.length}
            onChange={(event) => onFieldChange(event.target.value)}
          >
            <option value="">Select field</option>
            {fields.map((field) => (
              <option value={field.key} key={field.key}>
                {`${field.key} (${formatNumber(field.count)})`}
              </option>
            ))}
          </select>
        </label>
        <label>
          <span>Value</span>
          <input
            type="text"
            placeholder="value prefix"
            autoComplete="off"
            value={fieldFilterValue}
            disabled={!selectedFieldKey}
            onChange={(event) => onValueChange(event.target.value)}
          />
        </label>
        {fields.length ? (
          <div className="filterHint">
            {isActive ? "Matching field values by prefix." : "Choose a field and enter a value prefix."}
          </div>
        ) : (
          <div className="emptyHint">No fields loaded</div>
        )}
      </div>
    </section>
  );
}

function GcfFilter({
  items,
  selectedGcf,
  onToggle,
  onClear,
}: {
  items: GcfCount[];
  selectedGcf: string | null;
  onToggle: (value: string) => void;
  onClear: () => void;
}) {
  return (
    <section className="filterGroup">
      <div className="groupHeader">
        <h2>Top GCFs</h2>
        <button type="button" disabled={selectedGcf === null} onClick={onClear}>
          Clear
        </button>
      </div>
      <div className="gcfList">
        {items.length ? (
          items.map((item) => (
            <button
              type="button"
              className="gcfButton"
              aria-pressed={selectedGcf === item.value}
              key={item.value}
              onClick={() => onToggle(item.value)}
            >
              <span className="gcfValue">{item.value}</span>
              <span className="count">{formatNumber(item.count)}</span>
            </button>
          ))
        ) : (
          <div className="emptyHint">No gcf fields</div>
        )}
      </div>
    </section>
  );
}

function FilterGroup({
  title,
  items,
  selected,
  onToggle,
  onAll,
  onNone,
}: {
  title: string;
  items: [string, number][];
  selected: Set<string>;
  onToggle: (value: string) => void;
  onAll: () => void;
  onNone: () => void;
}) {
  return (
    <section className="filterGroup">
      <div className="groupHeader">
        <h2>{title}</h2>
        <div className="miniActions">
          <button type="button" onClick={onAll}>
            All
          </button>
          <button type="button" onClick={onNone}>
            None
          </button>
        </div>
      </div>
      <div className="chipGrid">
        {items.length ? (
          items.map(([label, count]) => (
            <Chip
              key={label}
              label={label}
              count={count}
              selected={selected.has(label)}
              onClick={() => onToggle(label)}
            />
          ))
        ) : (
          <div className="emptyHint">None</div>
        )}
      </div>
    </section>
  );
}

function Chip({
  label,
  count,
  selected,
  onClick,
}: {
  label: string;
  count: number;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button type="button" className="chip" aria-pressed={selected} onClick={onClick}>
      <span>{label}</span>
      <span className="count">{formatNumber(count)}</span>
    </button>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span className="metricLabel">{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function EventRow({
  event,
  previousEvent,
  selected,
  onSelect,
}: {
  event: ParsedEvent;
  previousEvent?: ParsedEvent;
  selected: boolean;
  onSelect: () => void;
}) {
  const timeGap =
    previousEvent && event.hasTime && previousEvent.hasTime
      ? `+${formatDuration(Math.abs(event.timeMs - previousEvent.timeMs))}`
      : "";
  const frameBadge = frameBadgeForEvent(event);
  const hasFrameValue = frameBadge.value !== null;

  return (
    <article
      className="eventRow"
      aria-selected={selected}
      role="button"
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onSelect();
        }
      }}
    >
      <div className="eventTime">
        <span>{event.timeText}</span>
        {timeGap ? <span>{timeGap}</span> : null}
      </div>
      <div className="frameCell">
        <span
          className={`frameBadge frameBadge-${hasFrameValue ? frameBadge.tone : "empty"}`}
          title={hasFrameValue ? `${frameBadge.field}=${frameBadge.value}` : "No cf/gcf"}
          aria-label={hasFrameValue ? `${frameBadge.field} ${frameBadge.value}` : "No cf or gcf"}
        >
          {frameBadge.value}
        </span>
      </div>
      <div className="eventBody">
        <div className="eventContent">
          <div>
            <div className="message">{event.message}</div>

            {event.fields.length ? (
              <div className="fieldList">
                {event.fields.slice(0, 10).map((field) => (
                  <span className="fieldPill" title={`${field.key}=${field.value}`} key={field.key}>
                    {`${field.key}=${field.value}`}
                  </span>
                ))}
                {event.fields.length > 10 ? (
                  <span className="fieldPill">{`+${event.fields.length - 10} fields`}</span>
                ) : null}
              </div>
            ) : null}
          </div>

          <div className="eventSide">
            <div className="eventMeta">
              <span className={`badge source-${event.source}`}>{event.source}</span>
              <span className={`badge level-${event.level.toLowerCase()}`}>{event.level}</span>
              <span className="fileLine">{`${event.fileName}:${event.lineNumber}`}</span>
            </div>
            <details className="rawDetails">
              <summary>Raw</summary>
              <pre>{event.parseError ? event.rawLine : JSON.stringify(event.raw, null, 2)}</pre>
            </details>
          </div>
        </div>
      </div>
    </article>
  );
}

function frameBadgeForEvent(event: ParsedEvent) {
  if (event.source === "server") {
    return { field: "gcf", value: getFieldValue(event, "gcf"), tone: "server" };
  }

  if (event.source === "client") {
    return { field: "cf", value: getFieldValue(event, "cf"), tone: "client" };
  }

  return { field: "frame", value: null, tone: "empty" };
}

async function collectLogFiles(directory: FileSystemDirectoryEntry): Promise<File[]> {
  const files: File[] = [];

  for await (const entry of directory.values()) {
    if (entry.kind === "file" && LOG_FILE_PATTERN.test(entry.name)) {
      files.push(await entry.getFile());
    }
  }

  return files.sort((a, b) => sourceRank(inferSource(a.name, {})) - sourceRank(inferSource(b.name, {})));
}

function parseSources(sources: LogSource[]): ParsedEvent[] {
  let ordinal = 0;
  const events: ParsedEvent[] = [];

  sources.forEach((source, sourceIndex) => {
    const lines = source.text.replace(/\r\n/g, "\n").split("\n");

    lines.forEach((line, lineIndex) => {
      const rawLine = line.trim();
      if (!rawLine) {
        return;
      }

      events.push(normalizeEvent({
        rawLine,
        fileName: source.name,
        lineNumber: lineIndex + 1,
        sourceIndex,
        ordinal,
      }));
      ordinal += 1;
    });
  });

  return events;
}

function normalizeEvent({
  rawLine,
  fileName,
  lineNumber,
  sourceIndex,
  ordinal,
}: {
  rawLine: string;
  fileName: string;
  lineNumber: number;
  sourceIndex: number;
  ordinal: number;
}): ParsedEvent {
  const parsed = parseJsonLine(rawLine);
  const record = parsed.ok ? parsed.value : parsePlainLine(rawLine);
  const timeValue = firstValue(record, TIME_KEYS);
  const parsedTime = parseTimeValue(timeValue, rawLine);
  const source = inferSource(fileName, record);
  const level = normalizeLevel(firstValue(record, LEVEL_KEYS));
  const message = normalizeMessage(firstValue(record, MESSAGE_KEYS), rawLine, parsed.ok);
  const fields = extractFields(record);
  const searchText = buildSearchText({
    fileName,
    lineNumber,
    source,
    level,
    message,
    fields,
    rawLine,
  });

  return {
    id: `${sourceIndex}:${lineNumber}:${ordinal}`,
    fileName,
    lineNumber,
    ordinal,
    rawLine,
    raw: record,
    parseError: !parsed.ok,
    source,
    level,
    message,
    timeText: parsedTime.text,
    timeMs: parsedTime.ms,
    hasTime: parsedTime.hasTime,
    fields,
    searchText,
  };
}

function parseJsonLine(rawLine: string): { ok: true; value: RawRecord } | { ok: false; value: null } {
  try {
    const value: unknown = JSON.parse(rawLine);
    if (value && typeof value === "object" && !Array.isArray(value)) {
      return { ok: true, value: value as RawRecord };
    }
  } catch {
    return { ok: false, value: null };
  }

  return { ok: false, value: null };
}

function parsePlainLine(rawLine: string): RawRecord {
  const record: RawRecord = { msg: rawLine };
  const timeMatch = rawLine.match(/\b\d{1,2}:\d{2}:\d{2}(?:\.\d{1,6})?\b/);
  const levelMatch = rawLine.match(/\b(TRACE|DEBUG|INFO|WARN|WARNING|ERROR|FATAL)\b/i);

  if (timeMatch) {
    record.time = timeMatch[0];
  }
  if (levelMatch) {
    record.level = levelMatch[1].toUpperCase();
  }

  return record;
}

function firstValue(record: RawRecord, keys: string[]): unknown {
  for (const key of keys) {
    if (Object.prototype.hasOwnProperty.call(record, key) && record[key] !== "") {
      return record[key];
    }
  }
  return undefined;
}

function parseTimeValue(value: unknown, rawLine: string): { hasTime: boolean; ms: number; text: string } {
  if (typeof value === "number" && Number.isFinite(value)) {
    const epochMs = value > 1_000_000_000_000 ? value : value * 1000;
    return {
      hasTime: true,
      ms: epochMs,
      text: formatEpochTime(epochMs),
    };
  }

  if (typeof value === "string" && value.trim()) {
    const text = value.trim();
    const timeOfDay = parseTimeOfDay(text);
    if (timeOfDay !== null) {
      return { hasTime: true, ms: timeOfDay, text: formatTimeOfDay(timeOfDay) };
    }

    const numeric = Number(text);
    if (Number.isFinite(numeric)) {
      const epochMs = numeric > 1_000_000_000_000 ? numeric : numeric * 1000;
      return {
        hasTime: true,
        ms: epochMs,
        text: formatEpochTime(epochMs),
      };
    }

    const parsed = Date.parse(text);
    if (Number.isFinite(parsed)) {
      return {
        hasTime: true,
        ms: parsed,
        text: formatEpochTime(parsed),
      };
    }
  }

  const fallbackMatch = rawLine.match(/\b\d{1,2}:\d{2}:\d{2}(?:\.\d{1,6})?\b/);
  if (fallbackMatch) {
    const timeOfDay = parseTimeOfDay(fallbackMatch[0]);
    if (timeOfDay !== null) {
      return { hasTime: true, ms: timeOfDay, text: formatTimeOfDay(timeOfDay) };
    }
  }

  return { hasTime: false, ms: Number.MAX_SAFE_INTEGER, text: "no time" };
}

function parseTimeOfDay(text: string): number | null {
  const match = text.match(/^(\d{1,2}):(\d{2}):(\d{2})(?:\.(\d{1,6}))?$/);
  if (!match) {
    return null;
  }

  const hours = Number(match[1]);
  const minutes = Number(match[2]);
  const seconds = Number(match[3]);
  if (hours > 23 || minutes > 59 || seconds > 59) {
    return null;
  }

  const fractional = (match[4] || "").padEnd(3, "0").slice(0, 3);
  const millis = fractional ? Number(fractional) : 0;
  return ((hours * 60 + minutes) * 60 + seconds) * 1000 + millis;
}

function parseContextRadius(text: string): number | null {
  const trimmed = text.trim();
  if (!/^\d+$/.test(trimmed)) {
    return null;
  }

  const value = Number(trimmed);
  return Number.isSafeInteger(value) ? value : null;
}

function formatEpochTime(epochMs: number): string {
  return new Date(epochMs).toLocaleString(undefined, { hour12: true });
}

function formatTimeOfDay(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const millis = ms % 1000;
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const suffix = hours >= 12 ? "PM" : "AM";
  const displayHours = hours % 12 || 12;
  const fractional = millis ? `.${String(millis).padStart(3, "0")}` : "";

  return `${displayHours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}${fractional} ${suffix}`;
}

function inferSource(fileName: string, record: RawRecord): string {
  const explicitSource = firstValue(record, SOURCE_KEYS);
  if (explicitSource !== undefined) {
    return normalizeSource(explicitSource);
  }

  const lowerName = fileName.toLowerCase();
  if (lowerName.includes("client")) {
    return "client";
  }
  if (lowerName.includes("server")) {
    return "server";
  }
  if (lowerName.includes("app")) {
    return "app";
  }
  if (lowerName.includes("garbage")) {
    return "garbage";
  }

  return normalizeSource(fileName.replace(/\.[^.]+$/, ""));
}

function normalizeSource(value: unknown): string {
  const text = String(value).trim();
  if (!text) {
    return "unknown";
  }
  return text.toLowerCase().replace(/[^a-z0-9_.:-]+/g, "-").replace(/^-+|-+$/g, "") || "unknown";
}

function normalizeLevel(value: unknown): string {
  if (value === undefined || value === null || value === "") {
    return "INFO";
  }
  return String(value).trim().toUpperCase();
}

function normalizeMessage(value: unknown, rawLine: string, parsedJson: boolean): string {
  if (value === undefined || value === null || value === "") {
    return parsedJson ? "(no message)" : rawLine;
  }
  if (typeof value === "string") {
    return value;
  }
  return stringifyValue(value);
}

function extractFields(record: RawRecord): LogField[] {
  return Object.entries(record)
    .filter(([key]) => !FIELD_EXCLUDE_KEYS.has(key))
    .map(([key, value]) => ({ key, value: stringifyValue(value) }));
}

function stringifyValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (value === null) {
    return "null";
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function buildSearchText(event: {
  fileName: string;
  lineNumber: number;
  source: string;
  level: string;
  message: string;
  fields: LogField[];
  rawLine: string;
}): string {
  return [
    event.fileName,
    String(event.lineNumber),
    event.source,
    event.level,
    event.message,
    ...event.fields.flatMap((field) => [field.key, field.value]),
    event.rawLine,
  ].join(" ").toLowerCase();
}

function parseSearchTerms(text: string): string[] {
  const terms: string[] = [];
  let current = "";
  let inQuote = false;

  function pushCurrent() {
    if (current) {
      terms.push(current.toLowerCase());
      current = "";
    }
  }

  for (const char of text.trim()) {
    if (char === "\"") {
      if (inQuote) {
        pushCurrent();
        inQuote = false;
      } else {
        pushCurrent();
        inQuote = true;
      }
      continue;
    }

    if (!inQuote && /\s/.test(char)) {
      pushCurrent();
      continue;
    }

    current += char;
  }

  pushCurrent();
  return terms;
}

function sortEvents(events: ParsedEvent[], sortDirection: SortDirection) {
  const direction = sortDirection === "desc" ? -1 : 1;
  events.sort((a, b) => {
    if (a.hasTime !== b.hasTime) {
      return a.hasTime ? -1 : 1;
    }

    const timeDelta = a.timeMs - b.timeMs;
    if (timeDelta !== 0) {
      return timeDelta * direction;
    }

    const sourceDelta = sourceRank(a.source) - sourceRank(b.source);
    if (sourceDelta !== 0) {
      return sourceDelta;
    }

    return a.ordinal - b.ordinal;
  });
}

function sourceRank(source: string): number {
  const index = SOURCE_ORDER.indexOf(source);
  return index === -1 ? SOURCE_ORDER.length : index;
}

function levelRank(level: string): number {
  const index = LEVEL_ORDER.indexOf(level);
  return index === -1 ? LEVEL_ORDER.length : index;
}

function countBy(events: ParsedEvent[], keyFn: (event: ParsedEvent) => string): Map<string, number> {
  const counts = new Map<string, number>();
  for (const event of events) {
    const key = keyFn(event);
    counts.set(key, (counts.get(key) || 0) + 1);
  }
  return counts;
}

function topGcfCounts(events: ParsedEvent[], limit: number): GcfCount[] {
  const counts = new Map<string, number>();
  for (const event of events) {
    const gcf = getFieldValue(event, "gcf");
    if (gcf !== null) {
      counts.set(gcf, (counts.get(gcf) || 0) + 1);
    }
  }

  return Array.from(counts.entries())
    .map(([value, count]) => ({ value, count }))
    .sort((a, b) => b.count - a.count || compareFieldValues(a.value, b.value))
    .slice(0, limit);
}

function eventHasFieldValue(event: ParsedEvent, key: string, value: string): boolean {
  return event.fields.some((field) => field.key === key && field.value === value);
}

function getFieldValue(event: ParsedEvent, key: string): string | null {
  const field = event.fields.find((item) => item.key === key && item.value.trim());
  return field ? field.value : null;
}

function compareFieldValues(a: string, b: string): number {
  const aNumber = Number(a);
  const bNumber = Number(b);
  if (Number.isFinite(aNumber) && Number.isFinite(bNumber) && aNumber !== bNumber) {
    return aNumber - bNumber;
  }
  return a.localeCompare(b);
}

function buildFieldIndex(events: ParsedEvent[]): FieldIndex {
  const fieldCounts = new Map<string, number>();
  const eventsByFieldValue = new Map<string, Map<string, ParsedEvent[]>>();

  for (const event of events) {
    const seenFields = new Set<string>();

    for (const field of event.fields) {
      if (!seenFields.has(field.key)) {
        fieldCounts.set(field.key, (fieldCounts.get(field.key) || 0) + 1);
        seenFields.add(field.key);
      }

      let valueIndex = eventsByFieldValue.get(field.key);
      if (!valueIndex) {
        valueIndex = new Map<string, ParsedEvent[]>();
        eventsByFieldValue.set(field.key, valueIndex);
      }

      const indexedEvents = valueIndex.get(field.value);
      if (indexedEvents) {
        indexedEvents.push(event);
      } else {
        valueIndex.set(field.value, [event]);
      }
    }
  }

  return {
    fields: Array.from(fieldCounts.entries())
      .map(([key, count]) => ({ key, count }))
      .sort((a, b) => b.count - a.count || a.key.localeCompare(b.key)),
    eventsByFieldValue,
  };
}

function applyFieldValueFilter(
  events: ParsedEvent[],
  fieldIndex: FieldIndex,
  selectedFieldKey: string,
  fieldFilterValue: string,
): ParsedEvent[] {
  const value = fieldFilterValue.trim();
  if (!selectedFieldKey || !value) {
    return events;
  }

  const valueIndex = fieldIndex.eventsByFieldValue.get(selectedFieldKey);
  if (!valueIndex) {
    return [];
  }

  const matches: ParsedEvent[] = [];
  const matchedIds = new Set<string>();
  for (const [indexedValue, indexedEvents] of valueIndex) {
    if (!indexedValue.startsWith(value)) {
      continue;
    }

    for (const event of indexedEvents) {
      if (!matchedIds.has(event.id)) {
        matches.push(event);
        matchedIds.add(event.id);
      }
    }
  }

  return matches;
}

function uniqueValues(events: ParsedEvent[], key: "source" | "level"): string[] {
  return Array.from(new Set(events.map((event) => event[key])));
}

function toggleSetValue(set: Set<string>, value: string): Set<string> {
  const next = new Set(set);
  if (next.has(value)) {
    next.delete(value);
  } else {
    next.add(value);
  }
  return next;
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60_000) {
    return `${(ms / 1000).toFixed(ms < 10_000 ? 2 : 1)}s`;
  }

  const minutes = Math.floor(ms / 60_000);
  const seconds = Math.round((ms % 60_000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatNumber(value: number): string {
  return numberFormatter.format(value);
}

function resultMeta(
  allEvents: ParsedEvent[],
  pageStart: number,
  pageEnd: number,
  visibleCount: number,
): string {
  if (!allEvents.length) {
    return "0 events";
  }

  const parseErrors = allEvents.filter((event) => event.parseError).length;
  const parsedText = parseErrors ? `, ${formatNumber(parseErrors)} plain-text lines` : "";
  const shownText =
    visibleCount > 0
      ? `${formatNumber(pageStart + 1)}-${formatNumber(pageEnd)} shown`
      : "0 shown";
  return `${shownText} of ${formatNumber(visibleCount)} visible${parsedText}`;
}

function contextResultMeta(
  allEvents: ParsedEvent[],
  selectedEvent: ParsedEvent | null,
  pageStart: number,
  pageEnd: number,
  visibleCount: number,
): string {
  if (!allEvents.length) {
    return "0 events";
  }
  if (!selectedEvent) {
    return "No selected line";
  }

  const shownText =
    visibleCount > 0
      ? `${formatNumber(pageStart + 1)}-${formatNumber(pageEnd)} shown`
      : "0 shown";
  return `${shownText} of ${formatNumber(visibleCount)} unfiltered around ${selectedEvent.fileName}:${selectedEvent.lineNumber}`;
}

function statusForLoadedFiles(sources: LogSource[]): string {
  const names = sources.map((source) => source.name).join(", ");
  return `Loaded ${sources.length} file${sources.length === 1 ? "" : "s"}: ${names}`;
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

export default App;
