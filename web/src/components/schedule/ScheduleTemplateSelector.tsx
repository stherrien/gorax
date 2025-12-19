import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Search, Loader2 } from 'lucide-react';
import { ScheduleTemplate } from '@/api/scheduleTemplates';
import { useScheduleTemplates } from '@/hooks/useScheduleTemplates';
import { ScheduleTemplateCard } from './ScheduleTemplateCard';

interface ScheduleTemplateSelectorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (template: ScheduleTemplate) => void;
}

const CATEGORIES = [
  { value: '', label: 'All Templates' },
  { value: 'frequent', label: 'Frequent' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'business', label: 'Business Hours' },
  { value: 'compliance', label: 'Compliance' },
  { value: 'security', label: 'Security' },
  { value: 'sync', label: 'Sync' },
  { value: 'monitoring', label: 'Monitoring' },
];

export function ScheduleTemplateSelector({
  open,
  onOpenChange,
  onSelect,
}: ScheduleTemplateSelectorProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const { data: templates, isLoading, error } = useScheduleTemplates({
    category: selectedCategory || undefined,
    search: searchQuery || undefined,
  });

  const handleSelect = (template: ScheduleTemplate) => {
    onSelect(template);
    onOpenChange(false);
  };

  const filteredTemplates = templates || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[85vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>Choose a Schedule Template</DialogTitle>
          <DialogDescription>
            Select a pre-configured schedule pattern or search for a specific template.
          </DialogDescription>
        </DialogHeader>

        {/* Search Bar */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
          <Input
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>

        {/* Category Tabs */}
        <Tabs value={selectedCategory} onValueChange={setSelectedCategory} className="flex-1 overflow-hidden flex flex-col">
          <TabsList className="grid grid-cols-5 lg:grid-cols-10 w-full">
            {CATEGORIES.map((category) => (
              <TabsTrigger key={category.value} value={category.value} className="text-xs">
                {category.label}
              </TabsTrigger>
            ))}
          </TabsList>

          {/* Loading State */}
          {isLoading && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="flex items-center justify-center py-12">
              <p className="text-sm text-destructive">
                Failed to load templates. Please try again.
              </p>
            </div>
          )}

          {/* Templates Grid */}
          {!isLoading && !error && (
            <TabsContent value={selectedCategory} className="flex-1 overflow-y-auto mt-4">
              {filteredTemplates.length === 0 ? (
                <div className="flex items-center justify-center py-12">
                  <p className="text-sm text-muted-foreground">
                    No templates found matching your criteria.
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pb-4">
                  {filteredTemplates.map((template) => (
                    <ScheduleTemplateCard
                      key={template.id}
                      template={template}
                      onSelect={handleSelect}
                      showDetails={true}
                    />
                  ))}
                </div>
              )}
            </TabsContent>
          )}
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
