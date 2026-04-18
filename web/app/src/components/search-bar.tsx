import { useState, useEffect, useCallback } from 'react'
import { Card, Input, Select } from '@hanzo/gui'
import { Search } from 'lucide-react'
import { getConfig } from '@/lib/config'

interface SearchBarProps {
  onSearch: (query: string) => void
  onFilterChange: (filter: string) => void
  onSortChange: (sort: string) => void
}

const filterOptions = [
  { label: 'None', value: 'none' },
  { label: 'Failing', value: 'failing' },
  { label: 'Unstable', value: 'unstable' },
]

const sortOptions = [
  { label: 'Name', value: 'name' },
  { label: 'Group', value: 'group' },
  { label: 'Health', value: 'health' },
]

export function SearchBar({ onSearch, onFilterChange, onSortChange }: SearchBarProps) {
  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState('none')
  const [sort, setSort] = useState('name')

  useEffect(() => {
    const cfg = getConfig()
    const storedFilter = localStorage.getItem('gatus:filter-by') || cfg.defaultFilterBy || 'none'
    const storedSort = localStorage.getItem('gatus:sort-by') || cfg.defaultSortBy || 'name'
    setFilter(storedFilter)
    setSort(storedSort)
    onFilterChange(storedFilter)
    onSortChange(storedSort)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleFilterChange = useCallback(
    (value: string) => {
      setFilter(value)
      localStorage.setItem('gatus:filter-by', value)
      onFilterChange(value)
    },
    [onFilterChange],
  )

  const handleSortChange = useCallback(
    (value: string) => {
      setSort(value)
      localStorage.setItem('gatus:sort-by', value)
      onSortChange(value)
    },
    [onSortChange],
  )

  return (
    <Card bordered padded className="flex flex-col gap-3 p-3 sm:p-4 lg:flex-row lg:gap-4">
      <div className="flex-1">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/50" />
          <Input
            placeholder="Search endpoints..."
            value={query}
            onChangeText={(text: string) => {
              setQuery(text)
              onSearch(text)
            }}
            className="w-full pl-10"
            size="$3"
          />
        </div>
      </div>
      <div className="flex flex-col gap-3 sm:flex-row sm:gap-3">
        <div className="flex flex-1 items-center gap-2 sm:flex-initial">
          <label className="whitespace-nowrap text-xs font-medium text-muted-foreground/70">
            Filter
          </label>
          <Select value={filter} onValueChange={handleFilterChange} size="$3">
            <Select.Trigger className="flex-1 sm:w-[120px]">
              <Select.Value />
            </Select.Trigger>
            <Select.Content>
              {filterOptions.map((o) => (
                <Select.Item key={o.value} value={o.value} index={filterOptions.indexOf(o)}>
                  <Select.ItemText>{o.label}</Select.ItemText>
                </Select.Item>
              ))}
            </Select.Content>
          </Select>
        </div>
        <div className="flex flex-1 items-center gap-2 sm:flex-initial">
          <label className="whitespace-nowrap text-xs font-medium text-muted-foreground/70">
            Sort
          </label>
          <Select value={sort} onValueChange={handleSortChange} size="$3">
            <Select.Trigger className="flex-1 sm:w-[100px]">
              <Select.Value />
            </Select.Trigger>
            <Select.Content>
              {sortOptions.map((o) => (
                <Select.Item key={o.value} value={o.value} index={sortOptions.indexOf(o)}>
                  <Select.ItemText>{o.label}</Select.ItemText>
                </Select.Item>
              ))}
            </Select.Content>
          </Select>
        </div>
      </div>
    </Card>
  )
}
