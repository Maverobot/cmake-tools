package cmakego

// clangConfigTemplate is a text template which will be added into CMakeLists.txt
const clangConfigTemplate = `
## ClangTools
include(${CMAKE_CURRENT_LIST_DIR}/cmake/ClangTools.cmake OPTIONAL
  RESULT_VARIABLE CLANG_TOOLS
)
if(CLANG_TOOLS)
  ${GLOB_SOURCE_SNIPPET}
  ${GLOB_HEADER_SNIPPET}
  add_format_target(${PROJECT_NAME} FILES ${SOURCES} ${HEADERS})
  add_tidy_target(${PROJECT_NAME}
    FILES ${SOURCES}
    DEPENDS ${TARGETS}
  )
endif()
`

// sourceSnippetTemplate defines how to find source files (*.cpp)
const sourceSnippetTemplate = `file(GLOB_RECURSE SOURCES
    $${GLOB_SOURCES}
  )`

// headerSnippetTemplate defines how to find header files (*.h)
const headerSnippetTemplate = `file(GLOB_RECURSE HEADERS
    $${GLOB_HEADERS}
  )`

// ConfigFileNames lists the names of config files and directories to be copied to the given cmake project
var configFileNames = [3]string{".clang-format", ".clang-tidy", "cmake"}
