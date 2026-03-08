import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
} from '@/components/ui/pagination'

export function PaginationSimple({ currentPage }: { currentPage: number }) {
  const totalPages = 5
  return (
    <Pagination>
      <PaginationContent>
        {Array.from({ length: totalPages }, (_, i) => {
          const pageNum = i + 1
          return (
            <PaginationItem key={pageNum}>
              <PaginationLink
                href={`?page=${pageNum}`}
                isActive={currentPage === pageNum}
              >
                {pageNum}
              </PaginationLink>
            </PaginationItem>
          )
        })}
      </PaginationContent>
    </Pagination>
  )
}
